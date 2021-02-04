import json
import sys

start_ts = None
end_ts = None

# start
ai_main_thread_used_cpu_sys_start = None
ai_main_thread_used_cpu_user_start = None
ai_queue_CPU_bthread_n1_used_cpu_total_start = None
ai_self_used_cpu_sys_start = None
ai_self_used_cpu_user_start = None

print(
    "Warning! This script was created to easily understand the first CPU bottleneck among main thread, BG thread, or backend.",
)
print(
    "Currently it only collects up to 1 background worker thread CPU usage. If your setup has more than one BG worker thread please adjust this script accordingly.",
)
with open(sys.argv[1]) as json_file:
    dd = json.load(json_file)
    server_stats = dd["ServerRunTimeStats"]
    print(
        "{},{},{},{},{}".format(
            "timeframe",
            "total_pct [0,#CORES]",
            "main_thread_pct [0,100]",
            "background thread [0,100]",
            "backend_pct",
        )
    )

    for ts, host_stats in server_stats.items():
        ai_main_thread_used_cpu_sys = 0
        ai_main_thread_used_cpu_user = 0
        ai_queue_CPU_bthread_n1_used_cpu_total = 0
        ai_self_used_cpu_sys = 0
        ai_self_used_cpu_user = 0
        for hostname, timeframe_server_runtime_stats_dict in host_stats.items():
            ai_main_thread_used_cpu_sys += float(
                timeframe_server_runtime_stats_dict["ai_main_thread_used_cpu_sys"]
            )
            ai_main_thread_used_cpu_user += float(
                timeframe_server_runtime_stats_dict["ai_main_thread_used_cpu_user"]
            )
            ai_queue_CPU_bthread_n1_used_cpu_total += float(
                timeframe_server_runtime_stats_dict[
                    "ai_queue_CPU_bthread_n1_used_cpu_total"
                ]
            )
            ai_self_used_cpu_sys += float(
                timeframe_server_runtime_stats_dict["ai_self_used_cpu_sys"]
            )
            ai_self_used_cpu_user += float(
                timeframe_server_runtime_stats_dict["ai_self_used_cpu_user"]
            )

        if ai_main_thread_used_cpu_sys_start is not None:
            ai_main_thread_used_cpu_sys_end = ai_main_thread_used_cpu_sys
            ai_main_thread_used_cpu_user_end = ai_main_thread_used_cpu_user
            ai_queue_CPU_bthread_n1_used_cpu_total_end = (
                ai_queue_CPU_bthread_n1_used_cpu_total
            )
            ai_self_used_cpu_sys_end = ai_self_used_cpu_sys
            ai_self_used_cpu_user_end = ai_self_used_cpu_user

            timeframe = (float(ts) - float(start_ts)) / 1000000000.0
            ai_main_thread_used_cpu_sys_pct = (
                (ai_main_thread_used_cpu_sys_end - ai_main_thread_used_cpu_sys_start)
                / timeframe
                * 100.0
            )
            ai_main_thread_used_cpu_user_pct = (
                (ai_main_thread_used_cpu_user_end - ai_main_thread_used_cpu_user_start)
                / timeframe
                * 100.0
            )
            main_thread_pct = (
                ai_main_thread_used_cpu_sys_pct + ai_main_thread_used_cpu_user_pct
            )
            ai_queue_CPU_bthread_n1_used_cpu_total_pct = (
                (
                    ai_queue_CPU_bthread_n1_used_cpu_total_end
                    - ai_queue_CPU_bthread_n1_used_cpu_total_start
                )
                / timeframe
                * 100.0
            )
            ai_self_used_cpu_sys_pct = (
                (ai_self_used_cpu_sys_end - ai_self_used_cpu_sys_start)
                / timeframe
                * 100.0
            )
            ai_self_used_cpu_user_pct = (
                (ai_self_used_cpu_user_end - ai_self_used_cpu_user_start)
                / timeframe
                * 100.0
            )
            backends_used_cpu_user_pct = (
                ai_self_used_cpu_user_pct - ai_main_thread_used_cpu_user_pct
            )
            backends_used_cpu_sys_pct = (
                ai_self_used_cpu_sys_pct - ai_main_thread_used_cpu_sys_pct
            )
            backend_pct = backends_used_cpu_user_pct + backends_used_cpu_sys_pct
            total_pct = ai_self_used_cpu_sys_pct + ai_self_used_cpu_user_pct
            print(
                "{},{},{},{},{}".format(
                    timeframe,
                    total_pct,
                    main_thread_pct,
                    ai_queue_CPU_bthread_n1_used_cpu_total_pct,
                    backend_pct,
                )
            )

        ai_main_thread_used_cpu_sys_start = ai_main_thread_used_cpu_sys
        ai_main_thread_used_cpu_user_start = ai_main_thread_used_cpu_user
        ai_queue_CPU_bthread_n1_used_cpu_total_start = (
            ai_queue_CPU_bthread_n1_used_cpu_total
        )
        ai_self_used_cpu_sys_start = ai_self_used_cpu_sys
        ai_self_used_cpu_user_start = ai_self_used_cpu_user
        start_ts = ts
