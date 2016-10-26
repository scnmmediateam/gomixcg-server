GOmixCG server and client for video mixing

Client can be found at https://github.com/sestg/gomixcg-client

Websocket commands:

Usage: command args data

COMMANDS:
update                //update command, needs args
  ARGS
    -clock 12 13 14 15  //updates clock with minutes seconds hundredths periods
    
graphics              //graphics command, needs args 
    -on                 //turns on graphics
    -off                //turns off graphics

vmix                  //vmix specific commands
caspar                //caspar specific commands
    -on                 //enable sending commands to the mixer
    -off                //disable sending commands to the mixer
    -ip 127.0.0.1       //set ip of the mixer
    -port 8081          //set port of the mixer
    -config              //lists config data
  
EG: vmix -ip 127.0.0.1 
    vmix -port 8081
    vmix -on
    

