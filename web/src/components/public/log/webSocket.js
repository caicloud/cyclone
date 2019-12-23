import PropTypes from 'prop-types';
import ReconnectingWebSocket from 'reconnecting-websocket';

const propTypes = {
  url: PropTypes.string.isRequired,
  onBefore: PropTypes.func,
  onData: PropTypes.func,
  onError: PropTypes.func,
};

const webSocket = props => {
  const { url, onBefore, onData, onError } = props;

  if (!url) {
    return;
  }

  const client = new ReconnectingWebSocket(url, null, {
    reconnectInterval: 3000,
  });
  client.addEventListener('open', () => {
    onBefore && onBefore();
  });
  client.addEventListener('message', e => {
    onData && onData(e.data);
  });
  client.addEventListener('error', e => {
    onError && onError(e);
  });

  if (!client.destroy) {
    client.destroy = client.close;
  }
  return client;
};

webSocket.propTypes = propTypes;

export default webSocket;
