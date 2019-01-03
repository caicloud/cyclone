import React from 'react';
import { Table, Button, Modal } from 'antd';
import { inject, observer } from 'mobx-react';
import IntegrationForm from './DataForm';
import PropTypes from 'prop-types';

@inject('integration')
@observer
class List extends React.Component {
  static propTypes = {
    integration: PropTypes.object,
  };
  state = { visible: false };
  componentDidMount() {
    this.props.integration.getIntegrationList();
  }
  addDataSource = () => {
    this.setState({
      visible: true,
    });
  };
  handleOk = e => {
    this.setState({
      visible: false,
    });
  };

  handleCancel = e => {
    this.setState({
      visible: false,
    });
  };
  render() {
    const { integration } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: intl.get('integration.type'),
        dataIndex: 'type',
        key: 'type',
      },
      {
        title: intl.get('integration.creationTime'),
        dataIndex: 'time',
        key: 'time',
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
        <Modal
          title={intl.get('integration.addexternalsystem')}
          visible={this.state.visible}
          footer={null}
          onCancel={this.handleCancel}
        >
          <IntegrationForm
            onSubmit={this.handleOk}
            onCancel={this.handleCancel}
          />
        </Modal>
      </div>
    );
  }
}

export default List;
