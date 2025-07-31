import { AxiosPromise } from "axios";
import request from "@/utils/request";
import {
  ConfigDto,
  ConfigsDto,
  ConfigUpdateDto,
  ConfigVo,
  Hysteria2AcmePathVo,
  Hysteria2ServerConfig,
} from "@/api/config/types";

export function getHysteria2ConfigApi(): AxiosPromise<Hysteria2ServerConfig> {
  return request({
    url: "/config/getHysteria2Config",
    method: "get",
  });
}

export function updateHysteria2ConfigApi(
  data: Hysteria2ServerConfig
): AxiosPromise {
  return request({
    url: "/config/updateHysteria2Config",
    method: "post",
    data: data,
  });
}

export function getConfigApi(data: ConfigDto): AxiosPromise<ConfigVo> {
  return request({
    url: "/config/getConfig",
    method: "get",
    params: data,
  });
}

export function listConfigApi(data: ConfigsDto): AxiosPromise<Array<ConfigVo>> {
  return request({
    url: "/config/listConfig",
    method: "post",
    data: data,
  });
}

export function updateConfigsApi(data: ConfigUpdateDto): AxiosPromise {
  return request({
    url: "/config/updateConfigs",
    method: "post",
    data: data,
  });
}

export function exportConfigApi(): AxiosPromise {
  return request({
    url: "/config/exportConfig",
    method: "post",
    responseType: "blob",
  });
}

export function importConfigApi(data: FormData): AxiosPromise {
  return request({
    url: "/config/importConfig",
    method: "post",
    headers: {
      "Content-Type": "multipart/form-data",
    },
    data: data,
  });
}

export function exportHysteria2ConfigApi(): AxiosPromise {
  return request({
    url: "/config/exportHysteria2Config",
    method: "post",
    responseType: "blob",
  });
}

export function importHysteria2ConfigApi(data: FormData): AxiosPromise {
  return request({
    url: "/config/importHysteria2Config",
    method: "post",
    headers: {
      "Content-Type": "multipart/form-data",
    },
    data: data,
  });
}

export function hysteria2AcmePathApi(): AxiosPromise<Hysteria2AcmePathVo> {
  return request({
    url: "/config/hysteria2AcmePath",
    method: "get",
  });
}

export function restartServerApi(): AxiosPromise {
  return request({
    url: "/config/restartServer",
    method: "post",
  });
}

export function uploadCertFileApi(data: FormData): AxiosPromise<string> {
  return request({
    url: "/config/uploadCertFile",
    method: "post",
    data,
    headers: { "Content-Type": "multipart/form-data" },
  });
}
// 第二节点相关API
export function getHysteria2Node2ConfigApi(): AxiosPromise<Hysteria2ServerConfig> {
  return request({
    url: "/config/getHysteria2Node2Config",
    method: "get",
  });
}

export function updateHysteria2Node2ConfigApi(
  data: Hysteria2ServerConfig
): AxiosPromise {
  return request({
    url: "/config/updateHysteria2Node2Config",
    method: "post",
    data: data,
  });
}

export function getSocks5ConfigApi(): AxiosPromise<Socks5ConfigVo> {
  return request({
    url: "/config/getSocks5Config",
    method: "get",
  });
}

export function updateSocks5ConfigApi(data: Socks5ConfigDto): AxiosPromise {
  return request({
    url: "/config/updateSocks5Config",
    method: "post",
    data: data,
  });
}

export function getNode2StatusApi(): AxiosPromise<Node2ConfigVo> {
  return request({
    url: "/config/getNode2Status",
    method: "get",
  });
}

export function toggleNode2Api(data: Node2ConfigDto): AxiosPromise {
  return request({
    url: "/config/toggleNode2",
    method: "post",
    data: data,
  });
}

export function getAllNodesStatusApi(): AxiosPromise<Record<string, boolean>> {
  return request({
    url: "/config/getAllNodesStatus",
    method: "get",
  });
}//
 第二节点导入导出API
export function exportNode2ConfigApi(): AxiosPromise {
  return request({
    url: "/config/exportNode2Config",
    method: "post",
    responseType: "blob",
  });
}

export function importNode2ConfigApi(data: FormData): AxiosPromise {
  return request({
    url: "/config/importNode2Config",
    method: "post",
    headers: {
      "Content-Type": "multipart/form-data",
    },
    data: data,
  });
}

export function exportFullConfigApi(): AxiosPromise {
  return request({
    url: "/config/exportFullConfig",
    method: "post",
    responseType: "blob",
  });
}