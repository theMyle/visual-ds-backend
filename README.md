## Techstack

- **Golang** 🥰🥰
  - **Goose** - clean, versioned database migrations
  - **SQLC** - type-safe query code generation
- **Supabase Postgres** - managed Postgres with smooth DX
- **Render** - simple, reliable backend deployment

## Tools

- **Draw.io** - quick visual schema planning
- **Curl & REST Client** - quick API request debugging

## Notes

> Hola! 🦖🦖

Backend runs on Render free tier with `512 RAM` and `0.1 CPU`. I chose golang specifically for this since it has very little overhead, is fast, and compiles to native code. With this, I can do my favorite part: `Optimization`. If I can turn that tiny amount of RAM into something that can handle a decent amount of requests without crashing, I'll feel very accomplished.

User experience with `Render` and `Supabase` is extremely good. Vey generous and gets the job done. Pair it with a `Nextjs` frontend and you have a fully free and modern full-stack application.

`SQLC` helped me a lot generating go boilerplates. I also enjoyed writing SQL and I do feel this is a much more superior flow that doing ORMs or just winging it with other DB management software, though I haven't really used anything else other than sqlyog. This gives you a lot of freedom and granularity, on top of that you don't waste much performance since these are direct query mappings.

I was blown away when I learned the you can stream HTTP responses. Like oh my goodness, HTTP runs on top of TCP. The usual flow was:

1. validate request
1. query from db
1. process
1. map data to DTO
1. then send response

I'm not really doing much processing so it goes directly to mapping. But that's the issue, data from the DB takes up space once inside the backend. What happens when that query returns `10k rows` or so, that might just crash the backend due to `OOM` (out of memory). And so I learned that in this particular scenario specially that most of the work is retriving data from the database. It can be improved from, it taking `O(n)` space to `O(1)` by simply doing the mapping and sending on the fly one by one, row by row directly without waiting for the whole thing. This poses a new problem though, what happens when you encounter and error halfway through? The client already received a `200 OK status` at the start when you sent the first byte.
