import { Table, Button } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from './MenuAction';

@inject('resource')
@observer
class List extends React.Component {
  static propTypes = {
    match: PropTypes.object,
    history: PropTypes.object,
    resource: PropTypes.object,
  };
  componentDidMount() {
    const { resource } = this.props;
    resource.listResourceTypes();
  }

  showModal = () => {
    const { match } = this.props;
    this.props.history.push(`${match.path}/add`);
  };

  render() {
    const { match, resource, history } = this.props;
    const columns = [
      {
        title: intl.get('resource.type'),
        dataIndex: 'spec.type',
        key: 'type',
      },
      {
        title: intl.get('resource.resolver'),
        dataIndex: 'spec.resolver',
        key: 'resolver',
      },
      {
        title: intl.get('resource.operations'),
        dataIndex: 'spec.operations',
        key: 'operations',
        render: value => _.join(value, ', '),
      },
      {
        title: intl.get('resource.binding'),
        dataIndex: 'spec.bind',
        key: 'binding',
        render: value => value.integrationType || '--',
      },
      {
        title: intl.get('creationTime'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'creationTime',
        render: value => FormatTime(value),
      },
      {
        title: intl.get('action'),
        dataIndex: 'spec.type',
        key: 'action',
        align: 'right',
        render: value => <MenuAction type={value} history={history} />,
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <Button type="primary" onClick={this.showModal}>
            {intl.get('operation.add')}
          </Button>
        </div>
        <Table
          columns={columns}
          rowKey={record => record.metadata.name}
          onRow={record => {
            return {
              onClick: () => {
                this.props.history.push(`${match.path}/${record.spec.type}`);
              },
            };
          }}
          dataSource={_.get(resource, 'resourceTypeList.items', [])}
        />
      </div>
    );
  }
}

export default List;
