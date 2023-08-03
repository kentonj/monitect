"""Send Images."""
import time
from typing import List
import websockets
import asyncio
import base64


HOST = "localhost:8080"
CAMERA_ID = "4ac03b16-b77b-4e55-9697-941077a0dd11"


async def post_image(websocket, image_path: str):
    # post an image to the API
    with open(image_path, 'rb') as f:
        await websocket.send(base64.b64encode(f.read()))


async def publish_image_frames(images: List[str], sleep_interval_s: int = 1):
    i = 0
    addr = f"ws://{HOST}/sensors/{CAMERA_ID}/publish"
    async with websockets.connect(addr) as websocket:
        while True:
            if i == len(images):
                i = 0
            await post_image(websocket, images[i])
            print(f'sent image: {images[i]}')
            print(f'here is the address: {addr}')
            i += 1
            time.sleep(sleep_interval_s)

async def publish_strings(strings: List[str]):
    i = 0
    addr = f"ws://{HOST}/sensors/{CAMERA_ID}/publish"
    async with websockets.connect(addr) as websocket:
        while True:
            if i == len(strings):
                i = 0
            s = strings[i]
            b = bytes(s, encoding='utf-8')
            print(f'sending bytes: {b}')
            await websocket.send(b)
            i += 1
            time.sleep(1)

images = ['pic1.png', 'pic2.png']
asyncio.run(publish_image_frames(images))

# strings = ['pic1.png', 'pic2.png']
# asyncio.run(publish_strings(strings))
