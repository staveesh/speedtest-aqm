import os
import subprocess
from pathlib import Path
from tqdm import tqdm
import time

bw_vals_mbps = [1000]
latency = [20]
loss = [0]
aqm = "cake"
# tools = ["ndt"]
interfaces = ["eth1"]
router_username = "root"
router_hostname = "192.168.1.1"
shaper_script = "aqm_shaper.sh"
server_ip = '192.168.1.166'

num_reps_per_config = 1

binary_path = "../../bin/bottleneck-finder"
output_root = "data/"

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

def run_iperf(ip):
    os.system(f"iperf3 -c {ip}")

def go(resume_index=0):
    
    print(f"##################### Run #{resume_index} Starting ########################")
    start_shaping(interfaces[0], bw_vals_mbps[0], latency[0], loss[0], aqm)
    time.sleep(1)

    check_aqm_status(interfaces[0])
    time.sleep(5)

    run_iperf(server_ip)

    output_dir = f"{output_root}/{interfaces[0]}_{bw_vals_mbps[0]}mbps_{latency[0]}ms_{loss[0]}loss_{aqm}"
    Path(output_dir).mkdir(parents=True, exist_ok=True)

    stop_shaping(interfaces[0], bw_vals_mbps[0], latency[0], loss[0], aqm)
    print(f"##################### Run #{resume_index} Ending ########################")

    time.sleep(1)

    check_if_stopped(interfaces[0])


# def run_tool(speedtest_tool, output_dir):
#     result = subprocess.run(
#         [binary_path, "-t", speedtest_tool, "-o", output_dir, "-a"],
#         stdout = subprocess.PIPE,
#         stderr = subprocess.PIPE,
#         text = True
#     )
#     print("############### Output #####################")
#     print(result.stdout)
#     print("############### Errors #####################")
#     print(result.stderr)
#     print("############################################")

if __name__ == '__main__':
    go()