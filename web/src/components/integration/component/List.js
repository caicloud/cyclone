import React from 'react';
import { Table, Button } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { IntegrationTypeMap } from '@/consts/const.js';

@inject('integration')
@observer
class List extends React.Component {
  static propTypes = {
    integration: PropTypes.object,
    history: PropTypes.object,
  };
  state = { visible: false };
  componentDidMount() {
    this.props.integration.getIntegrationList();
  }
  addDataSource = () => {
    this.props.history.push('/integration/addsource');
  };

  render() {
    const { integration } = this.props;
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
        dataIndex: 'spec',
        key: 'spec',
        render: spec => (
          <div>
            {_.get(spec, `${IntegrationTypeMap[spec.type]}.creationTime`)}
          </div>
        ),
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <h2>{intl.get('integration.datasource')}</h2>
          <Button type="primary" onClick={this.addDataSource}>
            {intl.get('operation.add')}
          </Button>
        </div>
        <Table columns={columns} dataSource={integration.integrationList} />
      </div>
    );
  }
}

export default List;
