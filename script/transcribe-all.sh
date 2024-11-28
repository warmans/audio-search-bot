#!/usr/bin/env sh

for inputFile in $*; do
  filename=$(basename -- "$inputFile");
  dirname=$(dirname -- "$inputFile")
  echo "Creating SRT: ${filename} -> ${filename%.*}.srt";
  ./bin/audio-search-bot transcribe mp3 --i "${inputFile}" --o "${dirname}/${filename%.*}.srt"
done;