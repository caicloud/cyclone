import axios from 'axios';
import http from '@/api/http';
import PropTypes from 'prop-types';

const propTypes = {
  url: PropTypes.string.isRequired,
  params: PropTypes.object,
  requestConfig: PropTypes.object,
  onBefore: PropTypes.func,
  onData: PropTypes.func,
  onComplete: PropTypes.func,
  onError: PropTypes.func,
};

const defaultProps = {
  requestConfig: {},
};

const network = props => {
  const { url, requestConfig, onBefore, onData, onError, onComplete } = props;

  if (!url) {
    return;
  }

  const CancelToken = axios.CancelToken;
  const source = CancelToken.source();

  onBefore && onBefore();
  http
    .get(url, requestConfig)
    .then(result => {
      onComplete && onComplete();
      onData && onData(result);
    })
    .catch(error => {
      onError && onError(error);
      onComplete && onComplete();
    });

  if (!source.destroy) {
    source.destroy = source.cancel;
  }
  return source;
};

network.propTypes = propTypes;
network.defaultProps = defaultProps;

export default network;
