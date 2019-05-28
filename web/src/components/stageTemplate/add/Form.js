import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import { Row, Col } from 'antd';
import { toJS } from 'mobx';
import FormContent from './FormContent';
import { validateForm } from './validate';
import styles from './template.module.less';

@inject('stageTemplate')
@observer
export default class StageTemplateForm extends React.Component {
  static propTypes = {
    history: PropTypes.object,
    match: PropTypes.object,
    stageTemplate: PropTypes.object,
    initialFormData: PropTypes.object,
    setTouched: PropTypes.func,
    isValid: PropTypes.bool,
    values: PropTypes.object,
  };

  componentDidMount() {
    const {
      match: { params },
      stageTemplate,
    } = this.props;
    const update = !!_.get(params, 'templateName');
    if (update) {
      stageTemplate.getTemplate(params.templateName);
    }
  }

  handleCancle = () => {
    const { history } = this.props;
    history.push('/stageTemplate');
  };

  mapRequestFormToInitForm = data => {
    const name = _.get(data, ['metadata', 'name'], '');
    const description = _.get(
      data,
      ['metadata', 'annotations', 'cyclone.dev/description'],
      ''
    );
    const scene = _.get(data, ['metadata', 'labels', 'cyclone.dev/scene'], '');
    const kind = _.get(
      data,
      ['metadata', 'labels', 'stage.cyclone.dev/template-kind'],
      ''
    );
    const spec = this.generateSpecObj(data);
    return {
      metadata: { name, description, scene, kind },
      spec,
    };
  };

  generateSpecObj = data => {
    const specData = _.get(data, 'spec', {});
    const containers = _.get(specData, 'pod.spec.containers', []);
    if (containers.length > 0) {
      containers.forEach(v => {
        v.command = _.last(v.command);
      });
    }
    let defaultSpec = {
      pod: {
        inputs: {
          arguments: [],
          resources: [],
        },
        outputs: {
          resources: [],
        },
        spec: {
          containers: [
            {
              env: [],
              image: '{{ image }}',
              command: ['{{{ cmd }}}'],
            },
          ],
        },
      },
    };
    return _.assign(defaultSpec, specData);
  };

  initFormValue = () => {
    const {
      match: { params },
    } = this.props;
    const update = !!_.get(params, 'templateName');
    if (update) {
      return this.mapRequestFormToInitForm(
        toJS(this.props.stageTemplate.template)
      );
    } else {
      return this.mapRequestFormToInitForm();
    }
  };

  generateData = data => {
    const metadata = {
      annotations: {
        'cyclone.dev/description': _.get(data, 'metadata.description', ''),
      },
      labels: {
        'cyclone.dev/scene': _.get(data, 'metadata.scene', ''),
        'stage.cyclone.dev/template-kind': _.get(data, 'metadata.kind', ''),
      },
      name: _.get(data, 'metadata.name', ''),
    };
    data.spec.pod.spec.containers.forEach(v => {
      v.command = _.concat(['/bin/sh', '-e', '-c'], v.command);
    });
    return { metadata, spec: data.spec };
  };

  submit = values => {
    const {
      stageTemplate,
      match: { params },
    } = this.props;
    const submitData = this.generateData(values);
    if (_.get(params, 'templateName')) {
      stageTemplate.updateStageTemplate(submitData, params.templateName, () => {
        this.props.history.replace('/stageTemplate');
      });
    } else {
      stageTemplate.createStageTemplate(submitData, () => {
        this.props.history.replace('/stageTemplate');
      });
    }
  };

  componentWillUnmount() {
    this.props.stageTemplate.resetTemplate();
  }

  render() {
    const {
      match: { params },
    } = this.props;
    const update = !!_.get(params, 'templateName');
    return (
      <div className={styles['stagetemplate-form']}>
        <div className="head-bar">
          <h2>
            {update ? intl.get('template.update') : intl.get('template.create')}
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
                  update={update}
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
