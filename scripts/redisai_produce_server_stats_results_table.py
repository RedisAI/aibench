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
            full_fname = "{}/{}".format(dirname, fname)
            with open(full_fname) as json_file:
                dd = json.load(json_file)
                workers = dd["Workers"]
                autobatching = dd["MetadataAutobatching"]
                tensorbatching = dd["TensorBatchSize"]
                rps = dd["OverallRates"]["overallOpsRate"]
                p50 = dd["OverallQuantiles"]["AllQueries"]["q50"]

                # we fix the tensor batch size to 1 for autobatching
                if tensorbatching == 1:
                    process_table_datapoint(autobatching, autobatching_arr, p50, workers, workers_arr,
                                            workers_autobatching_table_p50, full_fname)
                    process_table_datapoint(autobatching, autobatching_arr, rps, workers, workers_arr,
                                            workers_autobatching_table_rps, full_fname)
                #  we fix autobatching to 0 when doing tensor batching
                if autobatching == 0:
                    process_table_datapoint(tensorbatching, tensorbatching_arr, p50, workers, workers_arr,
                                            workers_tensorbatching_table_p50, full_fname)
                    process_table_datapoint(tensorbatching, tensorbatching_arr, rps, workers, workers_arr,
                                            workers_tensorbatching_table_rps, full_fname)

    workers_arr.sort()
    autobatching_arr.sort()
    tensorbatching_arr.sort()
    return workers_arr, autobatching_arr, workers_autobatching_table_rps, workers_autobatching_table_p50, tensorbatching_arr, workers_tensorbatching_table_rps, workers_tensorbatching_table_p50


def process_table_datapoint(metric_key, metric_arr, metric_value, workers, workers_arr, table, fname):
    metric_key_fname = "{}-fname".format(metric_key)
    if workers not in workers_arr:
        workers_arr.append(workers)
    if metric_key not in metric_arr:
        metric_arr.append(metric_key)
    if workers not in table:
        table[workers] = {}
    if metric_key not in table[workers]:
        table[workers][metric_key] = []
        table[workers][metric_key_fname] = []
    table[workers][metric_key].append(metric_value)
    table[workers][metric_key_fname].append(fname)


def print_results_table(workers_arr, metric_arr, metric_table, metric_str, functor=min,
                        print_last_server_runtime_stats=True, server_runtime_stats_metricname="used_memory_human"):
    print("Workers,{}".format(",".join(["{} {}".format(metric_str, x) for x in metric_arr])))
    for workersN in workers_arr:
        line = ["{} workers".format(workersN)]
        for metric_key in metric_arr:
            v = "n/a"
            metric_key_fname = "{}-fname".format(metric_key)
            if metric_key in metric_table[workersN]:
                v = functor(metric_table[workersN][metric_key])
                index = metric_table[workersN][metric_key].index(v)
                fname = metric_table[workersN][metric_key_fname][index]
                if print_last_server_runtime_stats:
                    runtime_stats_metric = "n/a"
                    with open(fname) as json_file:
                        dd = json.load(json_file)
                        server_runtime_stats = dd["ServerRunTimeStats"]
                        ts = list(server_runtime_stats.keys())
                        if len(ts) > 0:
                            last_stat_key = ts[-1]
                            first_host = list(server_runtime_stats[last_stat_key].keys())[0]
                            runtime_stats_metric = server_runtime_stats[last_stat_key][first_host][
                                server_runtime_stats_metricname]
                    v = '{}'.format(runtime_stats_metric)

            line.append(v)
        print(",".join([str(x) for x in line]))


parser = argparse.ArgumentParser(
    description="Simple script to process RedisAI results JSON and output overall metrics",
    formatter_class=argparse.ArgumentDefaultsHelpFormatter,
)
parser.add_argument("--dir", type=str, required=True)
parser.add_argument("--prefix", type=str, default="", help="prefix to filter the result files by")
parser.add_argument("--server_runtime_stats_metricname", type=str, default="used_memory_human",
                    help="The server runtime stat metric to extract from the last available datapoint per test")
args = parser.parse_args()

workers_arr, autobatching_arr, workers_autobatching_table_rps, workers_autobatching_table_p50, tensorbatching_arr, workers_tensorbatching_table_rps, workers_tensorbatching_table_p50 = process_json_files(
    args.dir, args.prefix)
print("-------------------")
print("Using the Overall inferences/sec to decide which result is the best per test variation")
print("-------------------")
print("## Auto-batching {} variation".format(args.server_runtime_stats_metricname))
print_results_table(workers_arr, autobatching_arr, workers_autobatching_table_rps, "Auto-batching", max, True,
                    args.server_runtime_stats_metricname)
print("")
print("-------------------")
print("## Tensor-batching {} variation".format(args.server_runtime_stats_metricname))
print_results_table(workers_arr, tensorbatching_arr, workers_tensorbatching_table_rps, "Tensor-batching", max, True,
                    args.server_runtime_stats_metricname)
print("")
