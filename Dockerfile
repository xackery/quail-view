FROM golang:1.21

RUN apt update && apt install -y xorg-dev libgl1-mesa-dev libopenal1 libopenal-dev libvorbis0a libvorbis-dev libvorbisfile3 make gcc-multilib  gcc-mingw-w64
WORKDIR /src