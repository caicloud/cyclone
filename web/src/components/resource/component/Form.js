import { Formik } from 'formik';
import { Modal, notification, Spin } from 'antd';
import FormContent from './FormContent';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';

@inject('project', 'resource')
@observer
class ResourceFrom extends React.Component {
  static propTypes = {
    visible: PropTypes.bool,
    modifyData: PropTypes.object,
    handleModalClose: PropTypes.func,
    update: PropTypes.bool,
    projectName: PropTypes.string,
    resource: PropTypes.object,
    project: PropTypes.object,
    resourceLen: PropTypes.number,
  };

  static defaultProps = {
    type: 'inputs',
  };

  constructor(props) {
    super(props);
    const {
      visible,
      resource: { listResourceTypes },
    } = props;
    this.state = {
      visible,
      loading: true,
      resourceTypeInfo: {},
    };
    listResourceTypes({}, data => {
      this.setResourceTypeInfo(data);
    });
  }

  setResourceTypeInfo = data => {
    const { modifyData } = this.props;
    const list = _.get(data, 'items', []);
    let argDes = {};
    const item = _.get(modifyData, 'spec.type')
      ? _.find(
          list,
          o => _.get(o, 'spec.type') === _.get(modifyData, 'spec.type')
        )
      : list[0];
    const arg = _.get(item, 'spec.parameters', []);
    _.forEach(arg, v => {
      argDes[v.name] = { description: v.description, required: v.required };
    });
    this.setState({
      resourceTypeInfo: {
        bindInfo: _.get(item, 'spec.bind'),
        arg,
        argDes,
      },
      loading: false,
    });
  };

  componentDidUpdate(preProps) {
    const { visible } = this.props;
    if (visible !== preProps.visible) {
      this.setState({ visible });
    }
  }

  closeModal = () => {
    const { handleModalClose } = this.props;
    this.setState({ visible: false });
    handleModalClose && handleModalClose(false);
  };

  getInitialValues = list => {
    const { modifyData, projectName, update, resourceLen } = this.props;
    const item = _.get(modifyData, 'spec.type')
      ? _.find(
          list,
          o => _.get(o, 'spec.type') === _.get(modifyData, 'spec.type')
        )
      : list[0];
    let data = {
      metadata: { name: '' },
      spec: {
        type: _.get(item, 'spec.type'),
        parameters: _.map(_.get(item, 'spec.parameters'), o =>
          _.pick(o, ['name', 'value'])
        ),
      },
    };
    if (!_.isEmpty(modifyData)) {
      data = _.merge(data, modifyData);
    }
    if (!update && !_.get(data, 'metadata.name')) {
      data.metadata.name = `${projectName}-rsc${resourceLen + 1}`;
    }
    return data;
  };

  submitResource = value => {
    const {
      handleModalClose,
      update,
      projectName,
      project: { createResource, updateResource },
    } = this.props;
    const { dirty } = this.state;
    const resourceObj = _.pick(value, ['metadata', 'spec']);
    if (update) {
      if (dirty) {
        updateResource(
          projectName,
          _.get(value, 'metadata.name'),
          resourceObj,
          () => {
            notification.success({
              message: intl.get('notification.updateResource'),
              duration: 2,
            });
            this.setState({ visible: false });
            handleModalClose(false);
          }
        );
      } else {
        this.setState({ visible: false });
        handleModalClose(false);
      }
    } else {
      createResource(projectName, resourceObj, () => {
        notification.success({
          message: intl.get('notification.createResource'),
          duration: 2,
        });
        this.setState({ visible: false });
        handleModalClose(false);
      });
    }
  };
  render() {
    const {
      update,
      resource: { resourceTypeList },
    } = this.props;
    const { visible, loading, resourceTypeInfo } = this.state;
    if (loading) {
      return <Spin />;
    }
    return (
      <Formik
        initialValues={this.getInitialValues(
          _.get(resourceTypeList, 'items', [])
        )}
        onSubmit={this.submitResource}
        render={props => (
          <Modal
            visible={visible}
            onCancel={this.closeModal}
            onOk={() => {
              this.setState({ dirty: props.dirty });
              props.handleSubmit();
            }}
          >
            <FormContent
              {...props}
              update={update}
              resourceTypeInfo={resourceTypeInfo}
            />
          </Modal>
        )}
      />
    );
  }
}

export default ResourceFrom;
