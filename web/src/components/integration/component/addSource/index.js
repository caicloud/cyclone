import React from 'react';
import IntegrationForm from './Form';
import { Row, Col } from 'antd';

export default class AddSource extends React.Component {
  render() {
    return (
      <div>
        <div className="head-bar">
          <h2>{intl.get('integration.addexternalsystem')}</h2>
        </div>
        <Row>
          <Col span={20}>
            <IntegrationForm onSubmit={this.handleOk} />
          </Col>
        </Row>
      </div>
    );
  }
}
