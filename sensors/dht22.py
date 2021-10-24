"""Taking measurements from the dht22"""
import os
import time
from datetime import datetime
from argparse import ArgumentParser
import requests

import adafruit_dht
import board

dht = adafruit_dht.DHT22(board.D4)


def sample() -> dict:
    # try to sample until a success
    max_retries = 5
    retry_counter = 0
    while True and retry_counter < max_retries:
        try:
            return {'temperature': dht.temperature, 'humidity': dht.humidity}
        except RuntimeError as e:
            print(f'sensor read error: {e}')
        time.sleep(1)
        retry_counter += 1
    # return an empty reading if it's not successful
    return None


def verify_sensor_registered(base_url: str, sensors: list, sensor_name: str) -> str:
    """Verify that the specified sensor name is in the list of returned sensors."""
    this_sensor = [x for x in sensors if x['name'] == sensor_name]
    if this_sensor:
        # get the sensor id
        sensor_id = this_sensor[0]['id']        
    else:
        # we need to register the sensor
        resp = requests.post(f"{base_url}/sensors", json={'name': sensor_name, 'type': 'thermometer'})
        resp.raise_for_status()
        sensor_id = resp.json()['sensor']['id']
    return sensor_id


def main():
    parser = ArgumentParser()
    parser.add_argument('--sensor-name', '-s', type=str, default=os.environ.get('SENSOR_NAME', 'dht22'))
    parser.add_argument('--server-url', type=str, default=os.environ.get('SERVER_URL', 'http://server:8181'))
    parser.add_argument('--interval', '-i', type=int, default=os.environ.get('INTERVAL', 300), help='sampling interval')
    args = parser.parse_args()

    sensor_name = args.sensor_name
    interval = args.interval
    base_url = args.server_url

    print(f'taking temperature for sensor {sensor_name} every {interval} seconds and posting to {base_url}')

    # verify that this sensor has been registered so far
    resp = requests.get(f"{base_url}/sensors")
    resp.raise_for_status()
    sensors = resp.json()['sensors']

    # verify a temperature and a humidity sensor
    temperature_sensor_id = verify_sensor_registered(base_url, sensors, f"{sensor_name}-temperature")
    humidity_sensor_id = verify_sensor_registered(base_url, sensors, f"{sensor_name}-humidity")

    # now start reading and posting results!
    while True:
        reading = sample()
        if reading:
            # reading is not none
            print(f'{sensor_name} reading @ {datetime.now()}: {reading}')
            resp = requests.post(f"{base_url}/sensors/{temperature_sensor_id}/readings", json={'value': reading['temperature']})
            if not resp.ok:
                print(f'something went wrong! {resp.status_code}')
            resp = requests.post(f"{base_url}/sensors/{humidity_sensor_id}/readings", json={'value': reading['humidity']})
            if not resp.ok:
                print(f'something went wrong! {resp.status_code}')
        time.sleep(args.interval)


if __name__ == '__main__':
    main()
