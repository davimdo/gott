# OTTPLAY

Tool to simulate an OTT player behabiour. Obataing the default Streams in the
manifest and requesting from the begining to the end.

Default Streams:

* Video: layer with the highest bitrate
* Audio: first stream in the manifest
* Subtitle: first stream in the mánifest


## Usage

```
usage: %s MANIFEST POSITION INTERVAL

	MANIFEST   OTT Manifest URL (for now Smooth)
	POSITION   Position in which start \"playing\"
	INTERVAL   Time interval in milliseconds to repeat the same manifest playout
```
