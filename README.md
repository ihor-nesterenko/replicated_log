# Replicated log iteration 2
First iteration of distributed systems homework.  
By default there is one main instance and two replicas.  
Feel free to modify compose.yaml with as many replicas as you want.  
To add more replicas put their address(host:port) in REPLICAS env var separated by comma 
## How to test  
1. docker compose build
2. docker compose up 
3. To view messages `curl localhost:8080/messages` (for child nodes change port to exposed one)
4. To add new message `curl -X POST localhost:8080/messages?concern=2 -d 'message'` (works only in main node). concern is a number of replicas we want to wait for
5. To add the delay add/update the env variable DELAY in compose.yaml. Delay is set in seconds
