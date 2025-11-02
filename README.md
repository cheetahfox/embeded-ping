# Long Ping 

Long Ping is a network observability tool that allows for high precision monitoring of long term network performance 
without requiring sending large number of packets for each sample. Long Ping does this by sending only 1 ping packet 
per second but keeping a rolling window of the last 15, 100 and 1000 packets to then calculate the network performance.
This makes Long Ping very similar to running a standard `ping` to a host but it can track long term trends and statistics. 

Currently Long Ping generates the following statistics for each host it's monitoring.

- Packets Sent
- Packets Received
- Duplicate Packets Received
- Packet Loss      : over the last 15, 100 and 1000 packets
- Average Latency  : over the last 15, 100 and 1000 packets
- Minimum Latency  : over the last 15, 100 and 1000 packets
- Maximum Latency  : over the last 15, 100 and 1000 packets
- Jitter           : over the last 15, 100 and 1000 packets