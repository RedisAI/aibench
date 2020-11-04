package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"

	"github.com/RedisAI/aibench/cmd/aibench_generate_data/common"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/fraud"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/serialize"
)

const (
	// Output data format choices (alphabetical order)
	formatRedisAI = "redisai"

	// Use case choices (make sure to update TestGetConfig if adding a new one)
	useCaseFraud = "creditcard-fraud"

	errTotalGroupsZero  = "incorrect interleaved groups configuration: total groups = 0"
	errInvalidGroupsFmt = "incorrect interleaved groups configuration: id %d >= total groups %d"

	defaultWriteSize = 4 << 20 // 4 MB
)

// semi-constants
var (
	formatChoices = []string{
		formatRedisAI,
	}
	useCaseChoices = []string{
		useCaseFraud,
	}
	// allows for testing
	fatal = log.Fatalf
)

// Program option vars:
var (
	format                         string
	useCase                        string
	profileFile                    string
	seed                           int64
	debug                          int
	interleavedGenerationGroupID   uint
	interleavedGenerationGroupsNum uint
	maxDataPoints                  uint64
	outputFileName                 string
	inputFileName                  string
)

// validateGroups checks validity of combination groupID and totalGroups
func validateGroups(groupID, totalGroupsNum uint) (bool, error) {
	if totalGroupsNum == 0 {
		// Need at least one group
		return false, fmt.Errorf(errTotalGroupsZero)
	}
	if groupID >= totalGroupsNum {
		// Need reasonable groupID
		return false, fmt.Errorf(errInvalidGroupsFmt, groupID, totalGroupsNum)
	}
	return true, nil
}

// validateFormat checks whether format is valid (i.e., one of formatChoices)
func validateFormat(format string) bool {
	for _, s := range formatChoices {
		if s == format {
			return true
		}
	}
	return false
}

// validateUseCase checks whether use-case is valid (i.e., one of useCaseChoices)
func validateUseCase(useCase string) bool {
	for _, s := range useCaseChoices {
		if s == useCase {
			return true
		}
	}
	return false
}

// GetBufferedWriter returns the buffered Writer that should be used for generated output
func GetBufferedWriter(fileName string) *bufio.Writer {
	// Prepare output file/STDOUT
	if len(fileName) > 0 {
		// Write output to file
		file, err := os.Create(fileName)
		if err != nil {
			fatal("cannot open file for write %s: %v", fileName, err)
		}
		return bufio.NewWriterSize(file, defaultWriteSize)
	}

	// Write output to STDOUT
	return bufio.NewWriterSize(os.Stdout, defaultWriteSize)
}

// Parse args:
func init() {

	flag.StringVar(&format, "format", "redisai", fmt.Sprintf("Format to emit. (choices: %s)", strings.Join(formatChoices, ", ")))

	flag.StringVar(&useCase, "use-case", "creditcard-fraud", fmt.Sprintf("Use case to model. (choices: %s)", strings.Join(useCaseChoices, ", ")))

	flag.IntVar(&debug, "debug", 0, "Debug printing (choices: 0, 1, 2). (default 0)")

	flag.UintVar(&interleavedGenerationGroupID, "interleaved-generation-group-id", 0,
		"Group (0-indexed) to perform round-robin serialization within. Use this to scale up data generation to multiple processes.")
	flag.UintVar(&interleavedGenerationGroupsNum, "interleaved-generation-groups", 1,
		"The number of round-robin serialization groups. Use this to scale up data generation to multiple processes.")

	flag.StringVar(&profileFile, "profile-file", "", "File to which to write go profiling data")
	flag.Int64Var(&seed, "seed", 0, "PRNG seed (default, or 0, uses the current timestamp).")

	flag.Uint64Var(&maxDataPoints, "max-transactions", 0, "Limit the number of transcactions to parse, 0 = no limit")
	flag.StringVar(&inputFileName, "input-file", "", "File name to read the data from")
	flag.StringVar(&outputFileName, "output-file", "", "File name to write generated data to")

	flag.Parse()

}

func main() {
	if ok, err := validateGroups(interleavedGenerationGroupID, interleavedGenerationGroupsNum); !ok {
		fatal("incorrect interleaved groups specification: %v", err)
	}
	if ok := validateFormat(format); !ok {
		fatal("invalid format specified: %v (valid choices: %v)", format, formatChoices)
	}
	if ok := validateUseCase(useCase); !ok {
		fatal("invalid use-case specified: %v (valid choices: %v)", useCase, useCaseChoices)
	}

	if len(profileFile) > 0 {
		defer startMemoryProfile(profileFile)()
	}

	rand.Seed(seed)

	// Get output writer
	out := GetBufferedWriter(outputFileName)
	defer func() {
		err := out.Flush()
		if err != nil {
			fatal(err.Error())
		}
	}()

	cfg := getConfig(useCase)
	sim := cfg.NewSimulator(maxDataPoints, inputFileName, debug)
	serializer := getSerializer(format)
	runSimulator(sim, serializer, out, interleavedGenerationGroupID, interleavedGenerationGroupsNum)
}

func runSimulator(sim common.Simulator, serializer serialize.TransactionSerializer, out io.Writer, groupID, totalGroups uint) {
	currGroupID := uint(0)
	point := serialize.NewTransaction()
	for !sim.Finished() {

		write := sim.Next(point)
		if !write {
			point.Reset()
			continue
		}

		// in the default case this is always true
		if currGroupID == groupID {
			err := serializer.Serialize(point, out)
			if err != nil {
				fatal("can not serialize point: %s", err)
				return
			}

		}
		point.Reset()

		currGroupID = (currGroupID + 1) % totalGroups
	}
}

func getConfig(useCase string) common.SimulatorConfig {
	switch useCase {
	case useCaseFraud:
		return &fraud.AibenchSimulatorConfig{
			InputFilename: outputFileName,
		}
	default:
		fatal("unknown use case: '%s'", useCase)
		return nil
	}
}

func getSerializer(format string) serialize.TransactionSerializer {
	switch format {
	case formatRedisAI:
		return &serialize.FraudSerializer{}
	default:
		fatal("unknown format: '%s'", format)
		return nil
	}
}

// startMemoryProfile sets up memory profiling to be written to profileFile. It
// returns a function to cleanup/write that should be deferred by the caller
func startMemoryProfile(profileFile string) func() {
	f, err := os.Create(profileFile)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}

	stop := func() {
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}

	// Catches ctrl+c signals
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c

		fmt.Fprintln(os.Stderr, "\ncaught interrupt, stopping profile")
		stop()

		os.Exit(0)
	}()

	return stop
}
