-- KEY[1]: 验证码的主键，例如 phone_code:login:13800138000
-- ARGV[1]: 用户输入的验证码
local key = KEYS[1]
local cntKey = key..":cnt"
local userCode = ARGV[1]

-- 步骤 1: 首先检查验证码是否存在。如果不存在，说明已过期或从未发送。
local correctCode = redis.call("get", key)
if not correctCode then
    return -3 -- 新增返回值 -3，代表“已过期或不存在”
end

-- 步骤 2: 检查剩余尝试次数。
local cnt = tonumber(redis.call("get", cntKey))
if cnt == nil or cnt <= 0 then
    return -1 -- 保持 -1，代表“尝试次数耗尽”
end

-- 步骤 3: 对比验证码。
if correctCode == userCode then
    -- 验证成功，立即删除 key，防止重复使用。
    redis.call("del", key, cntKey)
    return 0 -- 保持 0，代表“成功”
else
    -- 验证码错误，减少一次尝试次数。
    redis.call("decr", cntKey)
    return -2 -- 保持 -2，代表“验证码错误”
end
