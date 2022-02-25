# Roadmap

This roadmap details the planned features for Meteorae.

It is organized in three main sections, representing three major milestones of the project's life. The order and timeline is deliberately left open-ended, as the focus is more on the features themselves.

Each item is presented as a short user story, with sub-items representing the different tasks needed to be able to achieve that story.

If you want to take on part of the roadmap or have another idea you would like to work on, please [open a discussion]() to discuss your plan before opening a pull request.

## Minimum Usable Product

- [x] Users can log in
  - [x] Registering users
    - [ ] If not in setup mode, registering new users require authentication
  - [x] Generate JWT on log in
  - [x] Verify authentication on incoming requests
  - [x] Protect GraphQL methods
- [x] Users can add libraries with multiple root folders
  - [x] Using a GraphQL mutation to create a new library in DB
  - [x] Creating a DB queues up scanning jobs for all root directories
  - [x] The scanning process resolves files to items in the database
  - [x] The scanning process analyzes files to gather information
    - [x] Video support
    - [ ] Audio support
    - [x] Image support
    - [ ] Book support
- [ ] Users can play media
  - [x] Figure out a playback process
  - [ ] Generate playlists on the server
  - [ ] Media is served by the web server for playback
  - [ ] The server can figure out the best playback format for the user's session
  - [ ] Audio and video are transcoded with ffmpeg if the client cannot play them directly
    - [ ] Transcoded video and audio can be served using MPEG-DASH
    - [ ] Transcoded video and audio can be served using HLS
    - [ ] Direct files are served as `file.<ext>` to help clients know which format they are getting
    - [ ] Compatible subtitles are embedded in the MPEG-DASH Media Presentation Description or HLS Manifest
    - [ ] Incompatible subtitles are converted to a compatible format and are embedded in the MPEG-DASH Media Presentation Description or HLS Manifest
- [ ] Users can transcode any kind of supported files
  - [ ] Transcoders are chosen based on the type of file to be served
  - [ ] ffmpeg is used to transcode video and audio
  - [x] VIPS is used to transcode images
    - [ ] Local images should not be cached unless resized 
  - [ ] Archives are transcoded by serving their content using an index

## Minimum Viable Product

- [ ] Users can set permissions for accounts
  - [ ] Adding a library requires administrator permissions
- [ ] Files can be tagged with a hierarchical tag structure (Like: Parent / Child / Sub-Child)

## Minimum Lovable Product

- [ ] Users can automatically identify people in photos
  - [ ] Images can be linked to face coordinates
  - [ ] Image analysis generates face coordinates for pictures
  - [ ] Face coordinates can be linked to a person
  - [ ] The facial recognition engine can learn to recognize people based on existing mappings
    - [ ] Facial mappings with a lower confidence are added to a validation queue
    - [ ] Users are prompted to validate low confidence matches
- [ ] Users can find similar/duplicate media easily
  - [ ] During analysis, the server will generate a perceptual hash of the image or video
  - [ ] Similar media is added to a validation queue for processing by the administrator
- [ ] Users feels like the play queue is random
  - [ ] Shuffling uses something like Floyd-Steinberg instead of Fisher-Yates
  - [ ] Shuffling is reproducible based on a given seed
- [ ] Users can import existing tags
  - [ ] Can be imported from EXIF or XMP data
  - [ ] (Maybe) Can be imported from Hydrus Network tag repositories
  - [ ] (Maybe) Can be imported from existing web pages or services (Boorus, APIs, etc?)
