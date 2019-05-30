import { Spin, Table } from 'antd';
import { inject, observer } from 'mobx-react';
import Detail from '@/components/public/detail';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from '../MenuAction';

const { DetailHead, DetailHeadItem, DetailAction, DetailContent } = Detail;

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
        title: intl.get('name'),
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: intl.get('description'),
        dataIndex: 'description',
        key: 'description',
      },
    ];

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
        </DetailHead>
        <DetailContent>
          <Table
            columns={columns}
            dataSource={_.get(detail, 'spec.parameters')}
            pagination={false}
            rowKey="name"
          />
        </DetailContent>
      </Detail>
    );
  }
}

export default ResourceTypeDetail;
