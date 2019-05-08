import { Formik } from 'formik';
import { Modal, Button } from 'antd';
import BindResource from './BindResource';
import PropTypes from 'prop-types';
import { resourceParametersField } from '@/lib/const';

const Fragment = React.Fragment;

class ResourceFrom extends React.Component {
  static propTypes = {
    SetReasourceValue: PropTypes.func,
    type: PropTypes.oneOf(['input', 'ouput']),
  };

  static defaultProps = {
    type: 'input',
  };

  state = {
    visible: false,
  };

  getInitialValues = () => {
    const { type } = this.props;
    let data = {
      name: '',
      path: '',
      resourceType: 'SCM',
      spec: {
        parameters: [],
      },
    };
    if (type === 'input') {
      data.spec.parameters = resourceParametersField['SCM'];
    } else {
      data.spec.parameters = resourceParametersField['DockerRegistry'];
    }

    return data;
  };

  submitResource = value => {
    const { SetReasourceValue } = this.props;
    // TODO(qme): new add save data in store, then change stage resource field
    SetReasourceValue(value);
    this.setState({ visible: false });
  };
  render() {
    const { visible } = this.state;
    const { type } = this.props;
    return (
      <Formik
        initialValues={this.getInitialValues()}
        validate={() => {
          // TODO(qme): validate resource
        }}
        onSubmit={this.submitResource}
        render={props => (
          <Fragment>
            <Button
              ico="plus"
              onClick={() => {
                this.setState({ visible: true });
              }}
            >
              {intl.get('workflow.addResource')}
            </Button>
            <Modal
              visible={visible}
              onCancel={() => this.setState({ visible: false })}
              onOk={props.handleSubmit}
            >
              <BindResource {...props} type={type} />
            </Modal>
          </Fragment>
        )}
      />
    );
  }
}

export default ResourceFrom;
