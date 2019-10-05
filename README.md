# Itro
poc for playing midi through usb, recieving the audio via an interface, and Playing it

needs sudo in order to run
 

go build -o midirec .

sudo  DURATION=300 OP="udp" SLEEP=1 CHUNKS=512 SAVE=true  ./midirec 

