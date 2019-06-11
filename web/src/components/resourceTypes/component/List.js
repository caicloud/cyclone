import { Table, Button, Icon } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from './MenuAction';
import { Fragment } from 'react';

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
        render: (val, rowData) => {
          const builtin = _.get(rowData, [
            'metadata',
            'labels',
            'cyclone.dev/builtin',
          ]);
          return (
            <Fragment>
              {builtin && (
                <Icon
                  style={{ marginRight: '5px' }}
                  type="safety-certificate"
                  theme="twoTone"
                  twoToneColor="#1890ff"
                />
              )}
              <span>{_.get(rowData, 'spec.type', '')}</span>
            </Fragment>
          );
        },
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
        render: value => (value || {}).integrationType || '--',
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
        render: (value, rowData) => {
          const builtin =
            _.get(rowData, ['metadata', 'labels', 'cyclone.dev/builtin']) ===
            'true';
          return (
            <MenuAction type={value} disablAll={builtin} history={history} />
          );
        },
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
