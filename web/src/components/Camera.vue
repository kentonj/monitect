<template lang="pug">
.container
  .box
    .title {{ camera.name || '' }}
    .box
      img(:src="`data:image/png;base64,${imageBase64}`")
</template>

<script>
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
    this.connection = new WebSocket(`ws://localhost:8080/sensors/${this.$props.camera.id}/feed/read?client=frontend`);
    this.connection.onopen = function (event) {
      console.log(event);
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
