nohup ./AP_TLV1 > 1-`date '+%Y-%m-%d_%H'`.log  &
ps -ef |grep AP_TLV1|grep -v grep|awk -F ' ' '{print $2}'|xargs kill -9
