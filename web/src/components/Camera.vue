<template lang="pug">
.container
  .box
    .title {{ camera.name || '' }}
    .box
      img(:src="`data:image/jpeg;base64,${imageBase64}`")
</template>

<script>
import apiClient from '@/apiClient';

export default {
  name: 'Camera',
  props: {
    camera: Object,
  },
  data() {
    return {
      imageBase64: '',
      connection: null,
    };
  },
  created() {
    console.log('Starting connection to WebSocket Server');
    this.connection = apiClient.getSensorSocket(this.$props.camera.id, 'monitect-ui');
    this.connection.onopen = function () {
      console.log('Successfully connected to the websocket server...');
    };
  },
  mounted() {
    this.connection.onmessage = (event) => {
      this.imageBase64 = event.data;
    };
  },
};
</script>
