import IntegrationForm from './FormWrap';
import { Row, Col } from 'antd';
import { PropTypes } from 'mobx-react';

const CreateIntegration = props => {
  const {
    match: { params },
  } = props;
  const update = !!_.get(params, 'integrationName');

  return (
    <div className="integration-form">
      <div className="head-bar">
        <h2>
          {update
            ? intl.get('integration.updateexternalsystem')
            : intl.get('integration.addexternalsystem')}
        </h2>
      </div>
      <Row>
        <Col span={20}>
          <IntegrationForm {...props} />
        </Col>
      </Row>
    </div>
  );
};

CreateIntegration.propTypes = {
  match: PropTypes.object,
};

export default CreateIntegration;
