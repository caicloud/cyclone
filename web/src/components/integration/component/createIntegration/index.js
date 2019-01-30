import IntegrationForm from './FormWrap';
import { Row, Col } from 'antd';

const CreateIntegration = props => {
  return (
    <div>
      <div className="head-bar">
        <h2>{intl.get('integration.addexternalsystem')}</h2>
      </div>
      <Row>
        <Col span={20}>
          <IntegrationForm {...props} />
        </Col>
      </Row>
    </div>
  );
};

export default CreateIntegration;
