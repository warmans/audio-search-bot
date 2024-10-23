# Run server dependencies

local_resource(
    'audio-search-bot',
    dir='.',
    serve_dir='.',
    cmd='make build',
    serve_cmd='./bin/audio-search-bot bot',
    ignore=['./bin', './var', ".git"],
    deps='.',
    labels=['Bots'],
)
