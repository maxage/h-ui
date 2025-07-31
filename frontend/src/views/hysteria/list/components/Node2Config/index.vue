<template>
  <el-card shadow="never" style="margin-top: 20px;">
    <template #header>
      <div class="card-header">
        <span>{{ $t('hysteria.node2.title') }}</span>
        <el-switch 
          v-model="node2Config.enable" 
          @change="handleNode2EnableChange"
          :active-text="$t('hysteria.node2.enable')"
          :inactive-text="$t('hysteria.node2.disable')"
        />
      </div>
    </template>
    
    <div v-show="node2Config.enable">
      <!-- 第二节点基本配置 -->
      <el-form :model="node2Config" label-position="top">
        <el-tooltip :content="$t('hysteria.node2.remark')" placement="bottom">
          <el-form-item :label="$t('hysteria.node2.remark')" prop="remark">
            <el-input 
              v-model="node2Config.remark" 
              :placeholder="$t('hysteria.node2.remarkPlaceholder')"
              clearable 
            />
          </el-form-item>
        </el-tooltip>
        
        <el-form-item :label="$t('hysteria.node2.port')" prop="port">
          <el-input 
            v-model="node2Config.port" 
            :placeholder="$t('hysteria.node2.portPlaceholder')"
            disabled
          />
          <div class="form-item-tip">
            {{ $t('hysteria.node2.portTip') }}
          </div>
        </el-form-item>
        
        <el-form-item :label="$t('hysteria.node2.status')" prop="status">
          <el-tag 
            :type="node2Config.status ? 'success' : 'danger'"
            style="height: 32px"
          >
            {{ node2Config.status ? $t('hysteria.node2.running') : $t('hysteria.node2.stopped') }}
          </el-tag>
        </el-form-item>
      </el-form>
      
      <!-- SOCKS5 出站配置 -->
      <el-divider>{{ $t('hysteria.node2.socks5Config') }}</el-divider>
      
      <el-form :model="socks5Config" label-position="top" :rules="socks5Rules" ref="socks5FormRef">
        <el-tooltip :content="$t('hysteria.node2.socks5.addr')" placement="bottom">
          <el-form-item :label="$t('hysteria.node2.socks5.addr')" prop="addr">
            <el-input 
              v-model="socks5Config.addr" 
              :placeholder="$t('hysteria.node2.socks5.addrPlaceholder')"
              clearable 
            />
          </el-form-item>
        </el-tooltip>
        
        <el-tooltip :content="$t('hysteria.node2.socks5.username')" placement="bottom">
          <el-form-item :label="$t('hysteria.node2.socks5.username')" prop="username">
            <el-input 
              v-model="socks5Config.username" 
              :placeholder="$t('hysteria.node2.socks5.usernamePlaceholder')"
              clearable 
            />
          </el-form-item>
        </el-tooltip>
        
        <el-tooltip :content="$t('hysteria.node2.socks5.password')" placement="bottom">
          <el-form-item :label="$t('hysteria.node2.socks5.password')" prop="password">
            <el-input 
              v-model="socks5Config.password" 
              type="password"
              :placeholder="$t('hysteria.node2.socks5.passwordPlaceholder')"
              show-password
              clearable 
            />
          </el-form-item>
        </el-tooltip>
        
        <el-form-item>
          <el-button type="primary" @click="handleSaveSocks5Config">
            {{ $t('hysteria.node2.saveSocks5Config') }}
          </el-button>
          <el-button @click="handleTestSocks5Connection">
            {{ $t('hysteria.node2.testConnection') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </el-card>
</template>

<script lang="ts">
export default {
  name: "Node2Config",
};
</script>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { 
  getNode2StatusApi, 
  toggleNode2Api, 
  getSocks5ConfigApi, 
  updateSocks5ConfigApi 
} from '@/api/config';
import type { Node2ConfigDto, Socks5ConfigDto } from '@/api/config/types';

const socks5FormRef = ref();

const node2Config = reactive({
  enable: false,
  remark: 'Node2',
  port: 0,
  status: false
});

const socks5Config = reactive({
  addr: '',
  username: '',
  password: ''
});

const socks5Rules = {
  addr: [
    { required: true, message: 'SOCKS5地址不能为空', trigger: 'blur' },
    { pattern: /^.+:\d+$/, message: '请输入正确的地址格式 (host:port)', trigger: 'blur' }
  ]
};

// 加载第二节点状态
const loadNode2Status = async () => {
  try {
    const { data } = await getNode2StatusApi();
    node2Config.enable = data.enable;
    node2Config.remark = data.remark;
    node2Config.port = data.port;
    node2Config.status = data.status;
  } catch (error) {
    console.error('Failed to load node2 status:', error);
  }
};

// 加载SOCKS5配置
const loadSocks5Config = async () => {
  try {
    const { data } = await getSocks5ConfigApi();
    socks5Config.addr = data.addr;
    socks5Config.username = data.username;
    // 密码不从服务器返回，保持为空
  } catch (error) {
    console.error('Failed to load SOCKS5 config:', error);
  }
};

// 处理第二节点开关变化
const handleNode2EnableChange = async (enable: boolean) => {
  try {
    const dto: Node2ConfigDto = {
      enable,
      remark: node2Config.remark
    };
    
    await toggleNode2Api(dto);
    ElMessage.success(enable ? '第二节点已启用' : '第二节点已禁用');
    
    // 重新加载状态
    await loadNode2Status();
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败');
    // 恢复开关状态
    node2Config.enable = !enable;
  }
};

// 保存SOCKS5配置
const handleSaveSocks5Config = async () => {
  if (!socks5FormRef.value) return;
  
  try {
    await socks5FormRef.value.validate();
    
    const dto: Socks5ConfigDto = {
      addr: socks5Config.addr,
      username: socks5Config.username,
      password: socks5Config.password
    };
    
    await updateSocks5ConfigApi(dto);
    ElMessage.success('SOCKS5配置已保存');
    
    // 如果第二节点已启用，重新加载状态
    if (node2Config.enable) {
      await loadNode2Status();
    }
  } catch (error: any) {
    if (error.response?.data?.message) {
      ElMessage.error(error.response.data.message);
    } else {
      ElMessage.error('保存失败');
    }
  }
};

// 测试SOCKS5连接
const handleTestSocks5Connection = () => {
  // 这里可以添加测试连接的逻辑
  ElMessage.info('连接测试功能待实现');
};

onMounted(() => {
  loadNode2Status();
  loadSocks5Config();
});
</script>

<style lang="scss" scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-item-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>