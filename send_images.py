from datetime import datetime, timedelta, timezone
import requests
import base64

host = "http://localhost:8080"
camera_id = "01dcfa54-e11f-48e7-9db3-95592de43b77"
counter = 0

def post_image():
    # post an image to the API
    global counter
    pic_number = counter % 2 + 1
    resp = requests.post(f"{host}/sensors/{camera_id}/images", files={
        'image': open(f'pic{pic_number}.png', 'rb')
    })
    counter += 1


def get_latest():
    resp = requests.get(f"{host}/sensors/{camera_id}/images/latest")
    with open('latest_image.png', 'wb') as f:
        f.write(resp.content)


def truncate_images():
    oldest = datetime.now(timezone.utc) - timedelta(seconds=30)
    resp = requests.delete(f"{host}/sensors/{camera_id}/images", params={'oldest': oldest.isoformat()})
    print(resp.status_code)
    print(resp.json())


if __name__ == '__main__':
    while True:
        post_image()
    # truncate_images()
    get_latest()
