# Docker Image Generator
This tool allows you to create docker images or Dockerfiles customized with the tools of your choice (that are integrated in this tool).

This is ideal for when you need some tool that you don't have installed and/or don't want to install or mess your system with. Create your image, run the container, use the tool/s and throw it away when you are done, or keep it for future uses. 

## Installation

You will require to have installed [Docker](https://www.docker.com) and [go](https://golang.org) in your system. Eventually I will be releasing the binaries for the three major platforms, so you won't required to have go installed. 
Clone this repository and build the binary:

    git clone https://github.com/tuxotron/docker-image-generator
    cd docker-image-generator
    go build

## Usage

    ./doig
    This tool creates a customized docker image with the tools you need
    
    Usage: doig [--tools TOOLS] [--category CATEGORY] [--image IMAGE] [--dockerfile] [--list]
    
    Options:
      --tools TOOLS, -t TOOLS
                             List of tools separated by blank spaces
      --category CATEGORY, -c CATEGORY
                             List of categories separated by blank spaces
      --image IMAGE, -i IMAGE
                             Image name in lowercase
      --dockerfile, -d       Prints out the Dockerfile
      --list, -l             List the available tools and categories
      --help, -h             display this help and exit
      
## Examples

* List the tools available alphabetically:

        ./doig -l
        [*] Tools
          [-] altdns
          [-] amass
          [-] anonsurf
        ...
        [*] Categories
          [-] reversing
          [-] exploitation
          [-] osint
        ...
        
* Prints out a Dockerfile with the tools of your choice:

        ./doig -d -t nmap sqlmap wfuzz
        [*] Adding nmap
        [*] Adding sqlmap
        [*] Adding wfuzz
        
        FROM ubuntu:18.04
        
        RUN apt update && \
            apt install -y software-properties-common git curl p7zip-full locales && \
            add-apt-repository -y ppa:longsleep/golang-backports && \
            apt update && \
            localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8
        
        ENV LANG en_US.utf8
        
        RUN   apt install -y nmap && \
          apt install -y sqlmap && \
          apt install -y python-setuptools wfuzz && \
           rm -rf /var/lib/apt/lists/*
           
You can copy that file and use it with `docker [image] build`.

* Creates a docker image:

        ./doig -i mytools -t nmap sqlmap wfuzz
        [*] Adding nmap
        [*] Adding sqlmap
        [*] Adding wfuzz
        Step 1/4 : FROM ubuntu:18.04
         ---> 4e5021d210f6
         ...
         ---> 92bb29c4fc45
        Successfully built 92bb29c4fc45
        Successfully tagged mytools:latest
        
This creates a docker image called `mytools`. Now you can just get shell inside and use your tools:

        docker run -it --rm mytools
        root@dff16ee1f45c:/opt#

By default we'll be in /opt directory. Inside this directory you will find a file called tools.txt (if we added any tools when creating the image), which contains the list of the tools added to the image.
        
Or run a tool directly without getting a shell:

        docker run -it --rm mytools nmap localhost
        
        Starting Nmap 7.60 ( https://nmap.org ) at 2020-04-10 21:35 UTC
        Nmap scan report for localhost (127.0.0.1)
        ...
        
Each tool has assigned a category, so it is also possible to install all the tools of a specific category.

This would create an image with all the tools belonging to the `recon` category

        ./doig -i mytools -c recon
        
You can also spcify individual tools and categories. This would create an image with all the tools belonging to the `recon` category, plus sqlmap:

        ./doig -i mytools -c recon -t sqlmap

This would create an image with all tools:

        ./doig -i mytools -c all
        
## TO DO

This is in early stage, however it is functional. There are things to fix, to cleanup, improve and to add.

I hope you find this tool useful. 