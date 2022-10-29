<template lang="pug">
.container
  .box
    .title.is-3 {{ sensor.name || '' }}
    .box
      .title.is-4 {{ latestSensorReading.value }} ({{ sensor.unit || '' }})
      .subtitle.is-6 {{ latestSensorReading.createdTime }}
</template>

<script>
export default {
  name: 'Sensor',
  props: {
    sensor: Object,
  },
  data() {
    return {
      sensorReadings: [],
      latestSensorReading: {},
      connection: null,
    };
  },
  created() {
    console.log('Starting connection to WebSocket Server');
    this.connection = new WebSocket(`ws://localhost:8080/sensors/${this.$props.sensor.id}/feed/read?client=frontend`);
    this.connection.onmessage = function (event) {
      console.log(event);
    };
    this.connection.onopen = function (event) {
      console.log(event);
      console.log('Successfully connected to the echo websocket server...');
    };
  },
};
</script>
