import { Formik } from 'formik';
import { Modal } from 'antd';
import BindResource from './BindResource';
import PropTypes from 'prop-types';
import { resourceParametersField } from '@/lib/const';

class ResourceFrom extends React.Component {
  static propTypes = {
    SetReasourceValue: PropTypes.func,
    type: PropTypes.oneOf(['inputs', 'ouputs']),
    visible: PropTypes.bool,
    modifyData: PropTypes.object,
    handleModalClose: PropTypes.func,
  };

  static defaultProps = {
    type: 'inputs',
  };

  constructor(props) {
    super(props);
    const { visible } = props;
    this.state = {
      visible,
    };
  }

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

  getInitialValues = () => {
    const { type, modifyData } = this.props;
    let data = {
      name: '',
      path: '',
      resourceType: 'SCM',
      spec: {
        parameters: [],
      },
    };

    if (!_.isEmpty(modifyData)) {
      data = _.merge(data, modifyData);
    }
    if (type === 'inputs') {
      data.spec.parameters = resourceParametersField['SCM'];
    } else {
      data.spec.parameters = resourceParametersField['DockerRegistry'];
    }

    return data;
  };

  submitResource = value => {
    const { SetReasourceValue, handleModalClose } = this.props;
    // TODO(qme): new add save data in store, then change stage resource field
    SetReasourceValue(value);
    this.setState({ visible: false });
    handleModalClose(false);
  };
  render() {
    const { type } = this.props;
    const { visible } = this.state;
    return (
      <Formik
        initialValues={this.getInitialValues()}
        validate={() => {
          // TODO(qme): validate resource
        }}
        onSubmit={this.submitResource}
        render={props => (
          <Modal
            visible={visible}
            onCancel={this.closeModal}
            onOk={props.handleSubmit}
          >
            <BindResource {...props} type={type} />
          </Modal>
        )}
      />
    );
  }
}

export default ResourceFrom;
