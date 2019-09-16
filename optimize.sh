tunnel_path=$1
cd go && go build && ./go -tunnel-path=../$tunnel_path || true && 
cd .. && python3.7 discretizer.py -f $tunnel_path -d