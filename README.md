# music_server
Music server is a server which index musics on a specific folder (over network too).

h2. Developpment

Project in written in GO for server and rest services, and html5 / jQuery for interface.

Server index musics with id3tag info, cover (from MusicBrainz) and music length.

No database is required, work only with index files.

At home, it runs on a raspberry PI 2 and musics are on a SAN (Drobo).


h2. Compile

Dependences : 
* https://github.com/jotitan/go_embed_resources : to embed web resources in application. Can be used without go generate (call tool directly)
* https://github.com/mjibson/id3 : to extract id3tag info

I also use mp3info to extract real length of a music (in windows, must be install in index folder)

h2. Launch

To launch server : ./music_server with following arguments : 
* *-folder* _indexFolder_ 
* *_musicFolder_* : folder where musics will be.
* *-log* : folder where log must be written (/var/log for instance)
* *-port* : port of web application
* addressMask : if you want to launch musics update from web interface, control mask of requester IP
