"""Taking measurements from the dht22"""
import os
import time
from datetime import datetime
from argparse import ArgumentParser
import requests
import click
import dataclasses as dc
from sensors.util import get_sensor

import adafruit_dht
import board

SENSOR = adafruit_dht.DHT22(board.D4)


@dc.dataclass
class DHT22Reading:
    temperature: float
    humidity: float


def sample() -> DHT22Reading:
    # try to sample until a success
    max_retries = 5
    retry_counter = 0
    while True and retry_counter < max_retries:
        try:
            return DHT22Reading(SENSOR.temperature, SENSOR.humidity)
        except RuntimeError as e:
            print(f'sensor read error: {e}')
        time.sleep(1)
        retry_counter += 1
    # return an empty reading if it's not successful
    return None


def create_sensor(host: str, sensor_type: str, sensor_name: str, sensor_unit: str):
    resp = requests.post(f'{host}/api/sensors', json={
        'type': sensor_type,
        'name': sensor_name,
        'unit': sensor_unit
    })
    resp.raise_for_status()
    return resp.json()


@click.command()
@click.option('--host', type=str, default='http://rpi4.local:5000')
@click.option('--name', type=str, default='dht22')
@click.option('--interval', type=float, default=30)
def dht22(host: str, name: str, interval: float):
    # verify a temperature and a humidity sensor
    temperature_sensor = get_sensor(host, 'temperature', f'{name}-temperature')
    if temperature_sensor is None:
        temperature_sensor = create_sensor(host, 'temperature', f'{name}-temperature', 'C')
    temperature_sensor_id = temperature_sensor['id']
    print(f'using {temperature_sensor_id} for temperature sensor')
    humidity_sensor = get_sensor(host, 'humidity', f'{name}-humidity')
    if humidity_sensor is None:
        humidity_sensor = create_sensor(host, 'humidity', f'{name}-humidity', '%')
    humidity_sensor_id = humidity_sensor['id']
    print(f'using {humidity_sensor_id} for humidity sensor')
    # now start reading and posting results!
    while True:
        reading = sample()
        if reading:
            # reading is not none
            print(f'{name} reading @ {datetime.now()}: {reading}')
            resp = requests.post(f"{host}/api/sensors/{temperature_sensor_id}/readings", json={'value': reading.temperature})
            resp.raise_for_status()
            resp = requests.post(f"{host}/api/sensors/{humidity_sensor_id}/readings", json={'value': reading.humidity})
            resp.raise_for_status()
        time.sleep(interval)
