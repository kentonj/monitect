<template lang="pug">
.container
  .title {{ pongMessage }}
  .columns
    .column
      .box
        .title {{ camera.name }}
    .column
      Sensor(v-for="sensor in sensors"
      :key="sensor.id"
      :sensor="sensor"
      )
</template>

<script>
import Sensor from '@/components/Sensor.vue';

import axios from 'axios';

function getPong() {
  return axios.get('/api').then((response) => response.data);
}

function getCamera() {
  // get the first camera we found
  return axios.get('/api/sensors')
    .then((response) => response.data.sensors.find((sensor) => sensor.type === 'camera'));
}

function getSensors() {
  return axios.get('/api/sensors')
    .then((response) => response.data.sensors.filter((sensor) => sensor.type !== 'camera'));
}

export default {
  name: 'Dashboard',
  components: {
    Sensor,
  },
  data() {
    return {
      pongMessage: '',
      sensors: [],
      camera: {},
    };
  },
  methods: {
    setPong() {
      getPong().then((data) => {
        this.pongMessage = data.msg;
      });
    },
    setCamera() {
      getCamera().then((cam) => {
        console.log('this is the camera');
        console.log(cam);
        this.camera = cam;
      });
    },
    setSensors() {
      getSensors().then((sensors) => {
        this.sensors = sensors;
      });
    },
  },
  created() {
    this.setPong();
    this.setCamera();
    this.setSensors();
  },
};
</script>
