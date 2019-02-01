import KeyValue from '@/components/public/KeyValue';
import { Table } from 'antd';

import style from './detail.module.less';

const envColumns = [
  {
    title: intl.get('stage.spec.container.env.name'),
    dataIndex: 'name',
    width: '160px',
  },
  {
    title: intl.get('stage.spec.container.env.value'),
    dataIndex: 'value',
    width: '160px',
  },
];

const Configuration = ({ configuration = {} }) => {
  const containers = _.get(configuration, 'containers', []);
  return _.map(containers, (container, index) => (
    <div key={container.name || index}>
      <KeyValue
        cls={style['kv-item']}
        name={intl.get('stage.spec.container.image')}
        value={container.image}
      />
      <KeyValue
        cls={style['kv-item']}
        name={'EntryPoint'}
        isEmpty={_.isEmpty(container.command)}
        value={
          <div>
            {_.map(container.command, cmd => (
              <div key={cmd}>{cmd}</div>
            ))}
          </div>
        }
      />
      <KeyValue
        cls={style['kv-item']}
        name={'CMD'}
        isEmpty={_.isEmpty(container.args)}
        value={
          <div>
            {_.map(container.args, arg => (
              <div key={arg}>{arg}</div>
            ))}
          </div>
        }
      />
      <KeyValue
        cls={style['kv-item']}
        name={intl.get('stage.spec.container.envs')}
        isEmpty={_.isEmpty(container.env)}
        value={
          <Table
            columns={envColumns}
            dataSource={container.env}
            size="small"
            pagination={false}
          />
        }
      />
    </div>
  ));
};

export default Configuration;
