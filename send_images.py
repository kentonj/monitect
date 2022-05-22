from datetime import datetime, timedelta, timezone
import requests
import base64

host = "http://localhost:8080/"
camera_id = "fe3c89da-7ae5-417d-9a06-31cb73d66a42"


def post_image(image_path: str):
    # post an image to the API
    resp = requests.post(f"{host}/sensors/{camera_id}/images", files={
        'image': open(image_path, 'rb')
    })
    print(resp.status_code)


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
    post_image('web/src/assets/logo.png')
