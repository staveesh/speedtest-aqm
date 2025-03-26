import os
import subprocess
from pathlib import Path
from tqdm import tqdm
import time

bw_vals_mbps = [100, 200, 300, 400, 500, 600, 700, 800, 900]
latency = [10]
loss = [0]
aqm = ["no_aqm", "fq_codel", "codel", "sfq"]
# tools = ["ndt"]
interfaces = ["lan3"] # lan3 is download (hop 1), eth2 is upload (hop 2)
router_username = "root"
router_hostname = "192.168.1.1"
shaper_script = "aqm_shaper_nonetem.sh"
server_ip = '192.168.1.166:443'

num_reps_per_config = 1

binary_path = "../../bin/bottleneck-finder"
output_root = "data_json_iperf_new_highburst_nonetem/"

# get env variable KEY_PATH
key_path = os.environ.get("KEY_PATH")
if key_path is None:
    print("Please set the KEY_PATH environment variable to the path of the private key file.")
    exit(1)

def ssh_command(cmd):
    result = subprocess.run(
        ["ssh", "-i", f"{key_path}", f"{router_username}@{router_hostname}", cmd],
        stdout = subprocess.PIPE,
        stderr = subprocess.PIPE,
        text = True
    )
    print("############### Output #####################")
    print(result.stdout)
    print("############### Errors #####################")
    print(result.stderr)
    print("############################################")

def start_shaping(iface, bw, lat, loss, aqm):
    bw *= 1000
    cmd = f"/root/{shaper_script} start {iface} {bw} {lat} {loss} {aqm}"
    print(cmd)
    ssh_command(cmd)

def stop_shaping(iface, bw, lat, loss, aqm):
    bw *= 1000
    cmd = f"/root/{shaper_script} stop {iface} {bw} {lat} {loss} {aqm}"
    print(cmd)
    ssh_command(cmd)

def check_aqm_status(iface):
    # show ip link show dev eth1
    cmd = f"tc qdisc show dev {iface}"
    ssh_command(cmd)

def check_if_stopped(iface):
    # show ip link show dev eth1
    cmd = f"ip link show dev {iface}"
    ssh_command(cmd)

def run_iperf(ip, output_root):
    num_sessions = 5
    current_time = time.strftime("%Y%m%d-%H%M%S")
    file_name = f"iperf3_{current_time}.jsonl"
    output_dir = f"{output_root}/{file_name}"

    os.system(f"iperf3 -c {ip} -P {num_sessions} --json > {output_dir}")
    print(f"Iperf output saved to {output_dir}")

def run_speedtest(output_root):
    current_time = time.strftime("%Y%m%d-%H%M%S")
    file_name = f"speedtest_{current_time}.jsonl"
    output_dir = f"{output_root}/{file_name}"

    os.system(f"ndt7-client -server {server_ip} -no-verify -format json > {output_dir}")
    print(f"Speedtest output saved to {output_dir}")

def go(resume_index=0, iters=10):
   
    for bw_idx in range(len(bw_vals_mbps)):

        for aqm_idx in range(len(aqm)):

            print(f"##################### Run #{resume_index} Starting ########################")
            start_shaping(interfaces[0], bw_vals_mbps[bw_idx], latency[0], loss[0], aqm[aqm_idx])
            time.sleep(1)

            for idx in range(iters):
                print(f"##################### Run Iteration {idx} ########################")

                check_aqm_status(interfaces[0])
                time.sleep(1)

                output_dir = f"{output_root}/{interfaces[0]}_{bw_vals_mbps[bw_idx]}_{latency[0]}_{loss[0]}_{aqm[aqm_idx]}"
                Path(output_dir).mkdir(parents=True, exist_ok=True)

                run_iperf(server_ip.split(":")[0], output_dir)
                # run_speedtest(output_dir)

            stop_shaping(interfaces[0], bw_vals_mbps[bw_idx], latency[0], loss[0], aqm[aqm_idx])
            print(f"##################### Run #{resume_index} Ending ########################")

            time.sleep(1)

            check_if_stopped(interfaces[0])

if __name__ == '__main__':
    go()