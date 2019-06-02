import { Spin, Table, Tabs } from 'antd';
import { inject, observer } from 'mobx-react';
import Detail from '@/components/public/detail';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from '../MenuAction';

const { DetailHead, DetailHeadItem, DetailAction, DetailContent } = Detail;
const TabPane = Tabs.TabPane;

@inject('resource')
@observer
class ResourceTypeDetail extends React.Component {
  static propTypes = {
    match: PropTypes.object,
    resource: PropTypes.object,
    history: PropTypes.object,
  };

  constructor(props) {
    super(props);
    const {
      match: {
        params: { resourceTypeName },
      },
    } = this.props;
    this.props.resource.getResourceType(resourceTypeName);
  }

  getBinding = data => {
    const temp = [];
    _.forEach(data, (v, k) => {
      temp.push({
        name: k,
        binding: v,
      });
    });
    return temp;
  };

  render() {
    const {
      resource,
      match: {
        params: { resourceTypeName },
      },
      history,
    } = this.props;
    const loading = resource.resourceTypeLoading;
    if (loading) {
      return <Spin />;
    }
    const detail = resource.resourceTypeDetail;

    const columns = [
      {
        title: intl.get('parameterName'),
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: intl.get('required'),
        dataIndex: 'required',
        key: 'required',
        render: v => {
          return v ? 'true' : 'false';
        },
      },
      {
        title: intl.get('description'),
        dataIndex: 'description',
        key: 'description',
      },
    ];

    const bindingColumns = [
      {
        title: intl.get('parameterName'),
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: intl.get('resource.binding'),
        dataIndex: 'binding',
        key: 'binding',
      },
    ];

    const bindingData = this.getBinding(
      _.get(detail, 'spec.bind.paramBindings')
    );

    return (
      <Detail
        actions={
          <DetailAction>
            <MenuAction type={resourceTypeName} history={history} />
          </DetailAction>
        }
      >
        <DetailHead headName={_.get(detail, 'metadata.name')}>
          <DetailHeadItem
            name={intl.get('creationTime')}
            value={FormatTime(_.get(detail, 'metadata.creationTimestamp'))}
          />
          <DetailHeadItem
            name={intl.get('resource.type')}
            value={_.get(detail, 'spec.type')}
          />
          <DetailHeadItem
            name={intl.get('resource.operations')}
            value={_.get(detail, 'spec.operations') || ' -- '}
          />
          <DetailHeadItem
            name={intl.get('resource.binding')}
            value={_.get(detail, 'spec.bind.integrationType') || ' -- '}
          />
        </DetailHead>
        <DetailContent>
          <Tabs defaultActiveKey="parameters" type="card">
            <TabPane tab={intl.get('resource.parameters')} key="parameters">
              <Table
                columns={columns}
                dataSource={_.get(detail, 'spec.parameters')}
                pagination={false}
                rowKey="name"
              />
            </TabPane>
            <TabPane tab={intl.get('resource.binding')}>
              <Table
                columns={bindingColumns}
                dataSource={bindingData}
                pagination={false}
                rowKey="name"
              />
            </TabPane>
          </Tabs>
        </DetailContent>
      </Detail>
    );
  }
}

export default ResourceTypeDetail;
