import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import { Row, Col } from 'antd';
import FormContent from './FormContent';
import { IntegrationTypeMap } from '@/consts/const.js';
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

  generateData = data => {
    const metadata = {
      creationTime: Date.now().toString(),
      annotations: {
        'cyclone.dev/description': _.get(data, 'metadata.description', ''),
        'cyclone.dev/alias': _.get(data, 'metadata.alias', ''),
      },
    };
    const type = _.get(data, 'spec.type');
    const spec = _.pick(data.spec, [`${IntegrationTypeMap[type]}`, 'type']); // 只取type类型的表单
    if (type === 'SCM') {
      const scmValueMap = {
        UserPwd: ['user', 'password', 'type'],
        Token: ['token', 'server', 'type'],
      };
      const validateType = _.get(data, 'spec.scm.validateType');
      const scmObj = _.pick(spec.scm, scmValueMap[validateType]);
      spec[`${IntegrationTypeMap[type]}`] = scmObj;
    }

    if (type === 'Cluster') {
      const clusterValueMap = {
        UserPwd: ['user', 'password', 'server'],
        Token: ['bearerToken', 'server'],
      };
      const validateType = _.get(data, 'spec.cluster.credential.validateType');
      const clusterObj = _.pick(
        spec.cluster.credential,
        clusterValueMap[validateType]
      );
      const isControlCluster = _.get(data, 'spec.cluster.isControlCluster');
      const isWorkerCluster = _.get(data, 'spec.cluster.isWorkerCluster');
      spec[`${IntegrationTypeMap[type]}`] = {
        credential: clusterObj,
        tlsClientConfig: {
          insecure: true,
        },
        isControlCluster,
        isWorkerCluster,
      };
      if (isWorkerCluster) {
        const namespace = _.get(data, 'spec.cluster.namespace', '');
        const pvc = _.get(data, 'spec.cluster.pvc', '');
        _.assignIn(spec[`${IntegrationTypeMap[type]}`], { namespace, pvc });
      }
    }

    return { metadata, spec };
  };

  generateSpecObj = data => {
    let defaultSpec = {
      scm: {
        server: 'https://github.com',
        type: 'GitHub',
        validateType: 'Token',
      },
      dockerRegistry: {
        server: '',
        user: '',
        password: '',
      },
      sonarQube: {
        server: '',
        token: '',
      },
      cluster: {
        credential: {
          validateType: 'Token',
          server: '',
        },
        isControlCluster: false,
        isWorkerCluster: false,
      },
      type: '',
    };
    const type = _.get(data, 'spec.type');
    const spec = _.get(data, 'spec');
    const specData = _.pick(spec, [`${IntegrationTypeMap[type]}`, 'type']);
    if (type === 'SCM') {
      const token = _.get(data, 'spec.scm.token');
      if (!token) {
        specData.scm.validateType = 'UserPwd';
      } else {
        specData.scm.validateType = 'Token';
      }
    }
    if (type === 'Cluster') {
      const token = _.get(data, 'spec.cluster.credential.bearerToken');
      if (!token) {
        specData.cluster.credential.validateType = 'UserPwd';
      } else {
        specData.cluster.credential.validateType = 'Token';
      }
    }
    return _.assign(defaultSpec, specData);
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
    const creationTime = _.get(data, 'metadata.creationTimestamp', '');
    const spec = this.generateSpecObj(data);
    return {
      metadata: { alias, description, creationTime },
      spec,
    };
  };

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
