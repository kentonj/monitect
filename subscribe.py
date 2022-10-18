"""Subscribe to a websocket."""
import websockets
import asyncio

async def hello():
    async with websockets.connect("ws://localhost:8080/sensors/03017f18-da09-417e-8dc8-5d4109090b11/feed?clientId=kenton") as websocket:
        while True:
            res = await websocket.recv()
            print(res)

asyncio.run(hello())
