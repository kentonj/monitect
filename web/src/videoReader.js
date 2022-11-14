const mimeCodec = 'video/mp4; codecs="avc1.4D0033, mp4a.40.2"';
const queue = [];
let sourceBuffer = null;

// FIFO queue
function getEarliestFromQueue() {
  return queue.shift();
}

function addChunkToQueue(data) {
  queue.push(data);
}

// loadPacket gets the earliest element from the queue and adds it to the buffer
function addToBuffer() {
  if (!sourceBuffer.updating) {
    if (queue.length > 0) {
      sourceBuffer.appendBuffer(getEarliestFromQueue());
      console.log('added data from queue to buffer');
    } else {
      console.log('nothing in the queue');
    }
  } else {
    console.log('buffer is updating');
  }
}

function getMediaSource() {
  const ms = new MediaSource();
  ms.addEventListener('sourceopen', () => {
    sourceBuffer = ms.addSourceBuffer(mimeCodec);
    sourceBuffer.addEventListener('updateend', addToBuffer);
  });
  console.log('set the media source');
  return ms;
}

export default {
  getMediaSource,
  addChunkToQueue,
};
