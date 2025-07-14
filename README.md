# Long Ping 

Long Ping is a network observability tool that allows for high precision monitoring of long term network performance 
without requiring sending large number of packets for each sample. Long Ping does this by sending only 1 ping packet 
per second but keeping a rolling window of the last 15, 100 and 1000 packets to then calculate. 