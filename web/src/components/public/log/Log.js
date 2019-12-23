import { List } from 'antd';
import network from './network';
import webSocket from './webSocket';
import PropTypes from 'prop-types';
import styles from './index.module.less';

class Log extends React.Component {
  static propTypes = {
    url: PropTypes.string,
    queryParams: PropTypes.object,
    throttleWait: PropTypes.number,
    beforeConnected: PropTypes.func,
    requestConfig: PropTypes.object,
    parse: PropTypes.func,
  };
  static defaultProps = {
    throttleWait: 2000,
    total: 1000,
    showAllLog: false,
  };
  constructor(props) {
    super(props);
    this.state = {};
    this.logs = [];
    this.index = 1;
    this.logCoreInstance = React.createRef();
  }

  componentDidMount() {
    const { url, queryParams, throttleWait } = this.props;
    this.fetchData(url, queryParams);
    this.updateThrottle = _.throttle(this.updateData, throttleWait);
  }

  componentDidUpdate(prevProps) {
    const { url, queryParams } = this.props;
    const { url: prevUrl, queryParams: prevQueryParams } = prevProps;

    if (prevUrl !== url || !_.isEqual(prevQueryParams, queryParams)) {
      this.reset();
      this.fetchData(url, queryParams);
    }
  }

  componentWillUnmount() {
    this.destroyNetwork();
  }

  destroyNetwork = () => {
    if (this.fetchInstance && this.fetchInstance.destroy) {
      this.fetchInstance && this.fetchInstance.destroy();
    }
  };

  fetchData = (url, queryParams) => {
    if (!url) {
      return;
    }

    //TODO: destroy request instance first
    this.isWebSocket = _.startsWith(url, 'ws');
    if (this.isWebSocket) {
      this.fetchDataWithSocket(url);
    } else {
      this.fetchDataWithHttp(url, queryParams);
    }
  };

  fetchDataWithSocket = url => {
    const { beforeConnected } = this.props;
    this.fetchInstance = webSocket({
      url,
      onBefore: () => {
        this.setState({
          loading: false,
          continuous: true,
        });
        beforeConnected && beforeConnected();
      },
      onData: data => {
        this.parseData(data);
      },
    });
  };

  fetchDataWithHttp = (url, queryParams) => {
    if (this.locked) {
      return;
    }

    const { requestConfig, beforeConnected } = this.props;

    this.fetchInstance = network({
      url,
      requestConfig: requestConfig,
      onBefore: () => {
        this.locked = true;
        this.setState({
          loading: true,
          continuous: false,
        });
        beforeConnected && beforeConnected();
      },
      onComplete: () => {
        this.locked = false;
        this.setState({
          loading: false,
          continuous: false,
        });
      },
      onData: data => {
        this.parseData(data);
      },
    });
  };

  parseData = data => {
    const { parse } = this.props;
    if (parse) {
      const logData = parse(data);
      if (logData && _.isString(logData)) {
        this.logs.push(logData);
      } else if (_.isArray(logData)) {
        this.logs = this.logs.concat(logData);
      }
    } else {
      this.logs = this.logs.concat(data);
    }
    this.updateThrottle();
  };

  updateData = () => {
    const cacheLogs = _.map(this.logs, log => {
      return {
        key: this.index++,
        value: log,
      };
    });

    this.logs = [];

    this.setState(
      {
        data: cacheLogs,
      },
      () => {
        // TODO: steam log position
        // const bodyDOM = this.logCoreInstance.getBodyDOM();
        // if (this.isWebSocket) {
        //   // websocket 不断追加数据的时候，需要变更对去顶部/底部按钮的状态
        //   this.logCoreInstance.scrollButtonUpdate();
        //   // scroll 至最新的实时日志位置
        //   showLatestScreamLog && (bodyDOM.scrollTop = bodyDOM.scrollHeight);
        //   return;
        // }
        // const originValue = bodyDOM.scrollTop;
        // originValue !== 0 && (bodyDOM.scrollTop += offsetWhenDataChange);
        // // Less than one page, manually start loading the next page data.
        // bodyDOM.offsetHeight === bodyDOM.scrollHeight && this.update();
      }
    );
  };

  logContent = data => {
    return (
      <List
        dataSource={data}
        size="small"
        footer={null}
        header={null}
        bordered={false}
        split={false}
        renderItem={item => (
          <div className={styles['log-item']}>
            <div className="key">{_.get(item, 'key')}</div>
            <div className="value">{_.get(item, 'value')}</div>
          </div>
        )}
      />
    );
  };

  render() {
    const { data } = this.state;
    return (
      <div ref={this.logCoreInstance}>
        <div className={styles['log-body']}>{this.logContent(data)}</div>
      </div>
    );
  }
}

export default Log;
