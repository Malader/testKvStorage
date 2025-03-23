box.cfg{}

box.schema.space.create('kv', { if_not_exists = true })
box.space.kv:format({
    { name = 'key', type = 'string' },
    { name = 'value', type = 'map' },
})

box.space.kv:create_index('primary', { parts = { 'key' }, if_not_exists = true })

require('console').start()
