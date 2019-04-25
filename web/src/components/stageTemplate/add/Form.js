import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import { Row, Col } from 'antd';
import FormContent from './FormContent';
import { validateForm } from './validate';

@inject('stageTemplate')
@observer
export default class StageTemplateForm extends React.Component {
  constructor(props) {
    super(props);
    const {
      match: { params },
    } = props;
    this.update = !!_.get(params, 'integrationName');
    if (this.update) {
      props.integration.getIntegration(params.integrationName);
    }
  }

  static propTypes = {
    history: PropTypes.object,
    match: PropTypes.object,
    integration: PropTypes.object,
    initialFormData: PropTypes.object,
    setTouched: PropTypes.func,
    isValid: PropTypes.bool,
    values: PropTypes.object,
  };

  handleCancle = () => {
    const { history } = this.props;
    history.push('/integration');
  };

  generateData = data => {};

  submit = props => {};

  render() {
    return (
      <div className="integration-form">
        <div className="head-bar">
          <h2>
            {this.update
              ? intl.get('template.update')
              : intl.get('template.create')}
          </h2>
        </div>
        <Row>
          <Col span={20}>
            <Formik
              enableReinitialize={true}
              validate={validateForm}
              render={props => (
                <FormContent
                  {...props}
                  update={this.update}
                  submit={this.submit.bind(this, props)}
                  handleCancle={this.handleCancle}
                />
              )}
            />
          </Col>
        </Row>
      </div>
    );
  }
}
