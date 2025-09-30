-- 具体业务
-- 从调用方传入的第一个KEY，这是Redis中哈希（Hash）的键名。
-- 例如："interactive:article:123"
local key = KEYS[1]

-- 从调用方传入的第一个参数（ARGV），这是哈希里面的字段名（field）。
-- 例如："read_cnt", "like_cnt", "collect_cnt"
local cntKey = ARGV[1]

-- 从调用方传入的第二个参数，是要增加或减少的值（delta）。
-- tonumber() 将其从字符串转换为数字。例如：1 或 -1
local delta = tonumber(ARGV[2])

-- 检查这个哈希键是否存在于Redis中。
local exist=redis.call("EXISTS", key)

-- 如果键存在（exist 的结果为 1）
if exist == 1 then
    -- 就在这个哈希(key)中，为指定的字段(cntKey)增加(delta)值。
    -- HINCRBY 是一个原子操作，能保证并发安全。
    redis.call("HINCRBY", key, cntKey, delta)
    -- 返回 1，表示成功执行了增加操作。
    return 1
else
    -- 如果键不存在，则什么都不做。
    -- 返回 0，表示没有执行任何操作。
    return 0
end