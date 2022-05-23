from time import sleep
import io
import time
import requests
from time import sleep
from picamera import PiCamera
import click
from sensors.util import get_sensor

# set up camera
CAMERA = PiCamera()
CAMERA.resolution = (1024, 732)
CAMERA.start_preview()


@click.command()
@click.option('--host', type=str, default='http://rpi4.local:5000')
@click.option('--name', type=str, default='garage camera')
@click.option('--interval', type=float, default=5)
def capture(host: str, name: str, interval: float):
    camera_id = get_sensor(host, 'camera', name)['id']
    print(f'posting to camera: {camera_id}')
    stream = io.BytesIO()
    for frame in CAMERA.capture_continuous(stream, format='jpeg', use_video_port=True):
        stream.seek(0)
        resp = requests.post(f"{host}/api/sensors/{camera_id}/images", files={
            'image': stream
        })
        if resp.ok:
            image_id = resp.json()['imageId']
            print(f'status code: {resp.status_code} imageId: {image_id}')
        else:
            print(f'status code: {resp.status_code}')
        stream.truncate(0)
        stream.seek(0)
        time.sleep(interval)
