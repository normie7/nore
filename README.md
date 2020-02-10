# Noise Remover

#### Requisites

// requires

sudo apt-get install libmp3lame-dev

#### Short description
This project represents a web server with html interface for uploading and downloading your audio file and background 
service that cleans the audio file from background noise.

#### Project structure
##### configs
Folder contains config file with database credentials, assets folder and others

##### migrations
Folder with database migrations

##### web
Folder for web assets

`assets`
this folder is for web static assets such as css, img and js files

`templates`
this folder is for golang html templates

`temp-files`
this folder serves as a temporary storage for audio files. up folder is for uploaded files. And ready folder is for 
files that have been cleaned.

##### internal
Main folder of the application

`noiseremover` describes main business logic and defines two main interfaces - `Service` interface that allows to store
and retrieve audio files and to learn their current status. `BackgroundService` interface has methods to clean the audio
 files in the background. Also it defines repository and storage interfaces.
 
`api` the package describes http router for the application

`repository` package has repository implementation that stores data into the database (only MySQL at the moment)

`storage` is implementation of storage interface to work with files.

#### Roadmap

 - add tests
 - add comments
 - add pulse to the background service so every n minutes it would write "x files were cleaned in the last n minutes.
  y errors encountered."
 - add flags so that it could be possible to run web server and background server separately.
 - add cloud storage
 - add repository for postresql