kill -9 $(ps au|grep blogd|grep -v grep|awk -F ' ' '{print $2}')
nohup ./blogd > blogd.log &