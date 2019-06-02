import { Table, Button } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from './MenuAction';

@inject('integration')
@observer
class List extends React.Component {
  static propTypes = {
    integration: PropTypes.object,
    history: PropTypes.object,
    match: PropTypes.object,
  };
  state = { visible: false };
  componentDidMount() {
    this.props.integration.getIntegrationList();
  }
  addDataSource = () => {
    this.props.history.push('/integration/add');
  };
  render() {
    const { integration, history, match } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.name',
        key: 'name',
      },
      {
        title: intl.get('type'),
        dataIndex: 'spec.type',
        key: 'type',
      },
      {
        title: intl.get('integration.creationTime'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'spec',
        render: time => <div>{FormatTime(time)}</div>,
      },
      {
        title: intl.get('action'),
        dataIndex: 'metadata.name',
        key: 'action',
        align: 'right',
        render: value => <MenuAction name={value} history={history} />,
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <Button type="primary" onClick={this.addDataSource}>
            {intl.get('operation.add')}
          </Button>
        </div>
        <Table
          rowKey={record => record.metadata.name}
          columns={columns}
          onRow={record => {
            return {
              onClick: () => {
                this.props.history.push(
                  `${match.path}/${record.metadata.name}`
                );
              },
            };
          }}
          dataSource={integration.integrationList}
        />
      </div>
    );
  }
}

export default List;
