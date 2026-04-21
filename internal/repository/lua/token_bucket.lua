local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])

local data = redis.call("HMGET", key, "tokens", "timestamp")

local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    last_refill = now_ms
end

local elapsed_ms = now_ms - last_refill
local refill = (elapsed_ms * refill_rate) / 1000

tokens = math.min(capacity, tokens + refill)

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call("HMSET", key,
    "tokens", tokens,
    "timestamp", now_ms
)

return allowed
