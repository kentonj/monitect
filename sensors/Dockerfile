FROM python:3.7-slim-buster

RUN apt-get update
RUN apt-get install gcc libgpiod2 -y
RUN apt-get autoremove

COPY ./requirements.txt ./requirements.txt
RUN pip install -r ./requirements.txt

ENV PYTHONUNBUFFERED=1

COPY ./dht22.py ./dht22.py

ENTRYPOINT ["python"]

CMD ["dht22.py"]
