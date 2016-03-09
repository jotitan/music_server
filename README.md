# music_server
Music server is a server which index musics on a specific folder (over network too).

Project in written in GO for server and rest services, and html5 / jQuery for interface.

At home, it runs on a raspberry PI 2 and musics are on a SAN (Drobo).


Compile

Web resources are embedded in application.

Launch

To launch server : ./music_server -folder _indexFolder_ -musicFolder _musicFolder_ -log _logFolder_ 
