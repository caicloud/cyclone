import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import { Row, Col } from 'antd';
import { toJS } from 'mobx';
import FormContent from './FormContent';
import { validateForm } from './validate';
import { argumentsParamtersField } from '@/lib/const';

@inject('stageTemplate')
@observer
export default class StageTemplateForm extends React.Component {
  constructor(props) {
    super(props);
    const {
      match: { params },
    } = props;
    this.update = !!_.get(params, 'templateName');
    if (this.update) {
      props.stageTemplate.getTemplate(params.templateName);
    }
  }

  static propTypes = {
    history: PropTypes.object,
    match: PropTypes.object,
    stageTemplate: PropTypes.object,
    initialFormData: PropTypes.object,
    setTouched: PropTypes.func,
    isValid: PropTypes.bool,
    values: PropTypes.object,
  };

  handleCancle = () => {
    const { history } = this.props;
    history.push('/stageTemplate');
  };

  mapRequestFormToInitForm = data => {
    const alias = _.get(
      data,
      ['metadata', 'annotations', 'cyclone.dev/alias'],
      ''
    );
    const description = _.get(
      data,
      ['metadata', 'annotations', 'cyclone.dev/description'],
      ''
    );
    const spec = this.generateSpecObj(data);
    return {
      metadata: { alias, description },
      spec,
    };
  };

  generateSpecObj = data => {
    const specData = _.get(data, 'spec', {});
    let defaultSpec = {
      pod: {
        inputs: {
          arguments: argumentsParamtersField,
        },
        spec: {
          containers: [
            {
              env: [],
              image: '{{ image }}',
              args: ['/bin/sh", "-e", "-c", "{{{ cmd }}}'],
            },
          ],
        },
      },
    };
    return _.assign(defaultSpec, specData);
  };

  initFormValue = () => {
    const templateInfo = toJS(this.props.stageTemplate.template);
    return this.mapRequestFormToInitForm(templateInfo);
  };

  generateData = data => {
    const metadata = {
      annotations: {
        'cyclone.dev/description': _.get(data, 'metadata.description', ''),
        'cyclone.dev/alias': _.get(data, 'metadata.alias', ''),
      },
    };
    return { metadata, spec: data.spec };
  };

  submit = values => {
    const {
      stageTemplate,
      match: { params },
    } = this.props;
    const submitData = this.generateData(values);
    if (this.update) {
      stageTemplate.updateStageTemplate(submitData, params.templateName, () => {
        this.props.history.replace('/stageTemplate');
      });
    }
    stageTemplate.createStageTemplate(submitData, () => {
      this.props.history.replace('/stageTemplate');
    });
  };

  componentWillUnmount() {
    this.props.stageTemplate.resetTemplate();
  }

  render() {
    return (
      <div className="stagetemplate-form">
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
              initialValues={this.initFormValue()}
              enableReinitialize={true}
              validate={validateForm}
              onSubmit={this.submit}
              render={props => (
                <FormContent
                  {...props}
                  update={this.update}
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
