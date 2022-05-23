import click
from sensors.camera import capture
from sensors.dht22 import dht22

@click.group()
def cli():
    pass

cli.add_command(capture)
cli.add_command(dht22)

if __name__ == '__main__':
    cli()
