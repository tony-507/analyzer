"""
This script analyses SCTE-35 by reading output from ttt

* PAT and PMT change not supported
"""

import json
import os
import re
import sys
import pandas as pd

def get_preroll(scte35_fname: str, video_fname: str) -> float:
    """
    Calculate pre-roll of SCTE-35 message specified by the file <scte35_fname>
    using timing information from <video_fname>
    """
    splice_time = -1
    pkt_cnt = -1
    with open(scte35_fname, "rb") as f:
        scte35 = json.load(f)
        pkt_cnt = scte35["PktCnt"]
        if "SpliceCmd" in scte35:
            splice_time = scte35["SpliceCmd"]["SpliceTime"]
        else:
            raise Exception(f"Splice time not found in {scte35_fname}")
    
    v_info = pd.read_csv(video_fname)

    # Locate closest PES packet and regress PCR of SCTE-35 packet
    inf_pes_pkt = v_info[v_info["pktCnt"] < pkt_cnt].iloc[-1]
    sup_pes_pkt = v_info[v_info["pktCnt"] > pkt_cnt].iloc[0]

    slope = (sup_pes_pkt["pcr"] - inf_pes_pkt["pcr"]) / (sup_pes_pkt["pktCnt"] - inf_pes_pkt["pktCnt"])
    pkt_pcr = slope * (pkt_cnt - inf_pes_pkt["pktCnt"]) + inf_pes_pkt["pcr"]

    if splice_time != -1:
        idr_pcr = float(v_info[v_info["pts"] == splice_time]["pcr"].iloc[0])
    else:
        # Immediate
        idr_pcr = float(sup_pes_pkt["pcr"])

    return idr_pcr - pkt_pcr

if __name__ == "__main__":
    if len(sys.argv) < 4:
        print("Usage: python scte35Helper.py <output_dir> <video_pid> <scte35_pid>")
        exit(1)

    out_dir = sys.argv[1]
    video_pid = sys.argv[2]
    scte35_pid = sys.argv[3]
    scte35_cnt = 0

    for file in os.listdir(out_dir):
        r = re.match(f"{scte35_pid}_(.*)\.json", file)
        if r:
            # File index starts at 0
            scte35_cnt = max(scte35_cnt, int(r.group(1)) + 1)

    for idx in range(scte35_cnt):
        try:
            preroll = get_preroll(
                os.path.join(out_dir, f"{scte35_pid}_{idx}.json"),
                os.path.join(out_dir, f"{video_pid}.csv")
            )
            print(f"Preroll for {scte35_pid}_{idx} is {int(preroll / 27000)}ms")
        except:
            print(f"Fail to get preroll for {scte35_pid}_{idx}")
