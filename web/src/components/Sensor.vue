<template lang="pug">
.container
  .box
    .title.is-3 {{ sensor.name || '' }}
    .box
      .title.is-4 {{ latestSensorReading.value }} ({{ sensor.unit || '' }})
      //- .subtitle.is-6 {{ latestSensorReading.createdTime }}
</template>

<script>
import apiClient from '@/apiClient';

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
    this.connection = apiClient.getSensorSocket(this.$props.sensor.id, 'monitect-ui');
    this.connection.onopen = function () {
      console.log('Successfully connected to the websocket server...');
    };
  },
  mounted() {
    this.connection.onmessage = (event) => {
      console.log(`here is the event data: ${event.data}`);
      this.latestSensorReading = event.data;
    };
  },
};
</script>
