import argparse
import json
import os

def process_json_files(dirname: str, prefix: str = ""):
    workers_arr = []
    autobatching_arr = []
    tensorbatching_arr = []
    workers_autobatching_table_p50 = {}
    workers_autobatching_table_rps = {}
    workers_tensorbatching_table_p50 = {}
    workers_tensorbatching_table_rps = {}
    files_list = os.listdir(dirname)
    for fname in files_list:
        if ".json" in fname and ((prefix != "" and prefix in fname) or (prefix == "")):
            with open("{}/{}".format(dirname, fname)) as json_file:
                dd = json.load(json_file)
                workers = dd["Workers"]
                autobatching = dd["MetadataAutobatching"]
                tensorbatching = dd["TensorBatchSize"]
                rps = dd["OverallRates"]["overallOpsRate"]
                p50 = dd["OverallQuantiles"]["AllQueries"]["q50"]

                # we fix the tensor batch size to 1 for autobatching
                if tensorbatching == 1:
                    process_table_datapoint(autobatching, autobatching_arr, p50, workers, workers_arr,
                                            workers_autobatching_table_p50)
                    process_table_datapoint(autobatching, autobatching_arr, rps, workers, workers_arr,
                                            workers_autobatching_table_rps)
                #  we fix autobatching to 0 when doing tensor batching
                if autobatching == 0:
                    process_table_datapoint(tensorbatching, tensorbatching_arr, p50, workers, workers_arr,
                                            workers_tensorbatching_table_p50)
                    process_table_datapoint(tensorbatching, tensorbatching_arr, rps, workers, workers_arr,
                                            workers_tensorbatching_table_rps)

    workers_arr.sort()
    autobatching_arr.sort()
    tensorbatching_arr.sort()
    return workers_arr, autobatching_arr, workers_autobatching_table_rps, workers_autobatching_table_p50, tensorbatching_arr, workers_tensorbatching_table_rps, workers_tensorbatching_table_p50


def process_table_datapoint(metric_key, metric_arr, metric_value, workers, workers_arr, table):
    if workers not in workers_arr:
        workers_arr.append(workers)
    if metric_key not in metric_arr:
        metric_arr.append(metric_key)
    if workers not in table:
        table[workers] = {}
    if metric_key not in table[workers]:
        table[workers][metric_key] = []
    table[workers][metric_key].append(metric_value)


def print_results_table(workers_arr, metric_arr, metric_table, metric_str, functor=min):
    print("Workers,{}".format(",".join(["{} {}".format(metric_str, x) for x in metric_arr])))
    for workersN in workers_arr:
        line = ["{} workers".format(workersN)]
        for metric_key in metric_arr:
            v = "n/a"
            if metric_key in metric_table[workersN]:
                v = functor(metric_table[workersN][metric_key])
                v = '{:.3f}'.format(float(v))
            line.append(v)
        print(",".join([str(x) for x in line]))


parser = argparse.ArgumentParser(
    description="Simple script to process RedisAI results JSON and output overall metrics",
    formatter_class=argparse.ArgumentDefaultsHelpFormatter,
)
parser.add_argument("--dir", type=str, required=True)
parser.add_argument("--prefix", type=str, default="", help="prefix to filter the result files by")
args = parser.parse_args()

workers_arr, autobatching_arr, workers_autobatching_table_rps, workers_autobatching_table_p50, tensorbatching_arr, workers_tensorbatching_table_rps, workers_tensorbatching_table_p50 = process_json_files(
    args.dir, args.prefix)

print("-------------------")
print("## Auto-batching overall throughput (inferences/sec) ((higher is better))")
print_results_table(workers_arr, autobatching_arr, workers_autobatching_table_rps, "Auto-batching",max)
print("")
print("## Auto-batching p50 latency results (latency in ms including RTT) ((lower is better))")
print_results_table(workers_arr, autobatching_arr, workers_autobatching_table_p50, "Auto-batching",min)
print("")
print("-------------------")
print("## Tensor-batching overall throughput (inferences/sec) ((higher is better))")
print_results_table(workers_arr, tensorbatching_arr, workers_tensorbatching_table_rps, "Tensor-batching",max)
print("")
print("## Tensor-batching p50 latency results (latency in ms including RTT) ((lower is better))")
print_results_table(workers_arr, tensorbatching_arr, workers_tensorbatching_table_p50, "Tensor-batching",min)
print("")