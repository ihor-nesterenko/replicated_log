# Replicated log iteration 1  
First iteration of distributed systems homework.  
By default there is one main instance and two replicas.  
Feel free to modify compose.yaml with as many replicas as you want.  
To add more replicas put their address(host:port) in REPLICAS env var separated by comma 
## How to test  
1. docker compose build
2. docker compose up 
3. To view messages `curl localhost:8080/messages` (for child nodes change port to exposed one)
4. To add new message `curl -X POST localhost:8080/messages -d 'message'` (works only in main node)
