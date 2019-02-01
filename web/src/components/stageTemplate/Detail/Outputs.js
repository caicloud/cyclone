import { Table, Collapse } from 'antd';
import PropTypes from 'prop-types';

const Outputs = ({ outputs = {} }) => {
  const artifactColumns = [
    {
      title: intl.get('name'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: intl.get('path'),
      dataIndex: 'path',
    },
  ];

  return (
    <Collapse activeKey={['1']}>
      <Collapse.Panel
        showArrow={false}
        header={intl.get('stage.outputs')}
        key="1"
      >
        <Table
          columns={artifactColumns}
          dataSource={outputs.artifacts}
          pagination={false}
          rowKey="name"
        />
      </Collapse.Panel>
    </Collapse>
  );
};

Outputs.propTypes = {
  outputs: PropTypes.shape({
    artifacts: PropTypes.arrayOf(
      PropTypes.shape({
        name: PropTypes.string,
        path: PropTypes.string,
      })
    ),
  }).isRequired,
};

export default Outputs;
