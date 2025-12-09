# Requirements
-> Realtime updates required
-> Log file is very long might be in GB (need last 10 lines)


# My thinking
-> As realtime updates are required, we can eiter choose between SSE or websockets
-> using SSE, as this is a one to many case, we do not require bidirectional communication
-> language preferred -> golang (good for these use cases, more familiarity)
