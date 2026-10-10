import { createVrpcClient, VRPC_JSON_CONTENT_TYPE, type VrpcTransportRequest } from '@yorun-ai/vrpc';
import {
  createOrderApiService,
  createOrderApiServiceV1,
  createOrderApiServiceV2,
  createOrderApiServiceV10,
  type Order,
} from './generated';

export async function checkVersionedClients() {
  const requests: VrpcTransportRequest[] = [];
  const client = createVrpcClient({
    prefixUrl: 'https://example.invalid/invoke',
    clientInfo: {
      clientName: 'skelc.ci',
      clientVersion: '1.0.0',
      clientInstanceId: '123e4567-e89b-42d3-a456-426614174000',
    },
    transport: {
      async request(request) {
        requests.push(request);
        const { params } = JSON.parse(String(request.body));
        return {
          status: 200,
          headers: new Headers({
            'content-type': VRPC_JSON_CONTENT_TYPE,
            'vrpc-status': 'OK',
            'vrpc-server': 'name=skelc.fixture,version=1.0.0,instanceId=123e4567-e89b-42d3-a456-426614174001',
          }),
          body: new TextEncoder().encode(JSON.stringify({ result: { id: params.id } })),
          url: request.url,
        };
      },
    },
  });

  const services = [
    ['', createOrderApiService(client)],
    ['V1', createOrderApiServiceV1(client)],
    ['V2', createOrderApiServiceV2(client)],
    ['V10', createOrderApiServiceV10(client)],
  ] as const;
  for (const [index, [version, service]] of services.entries()) {
    const id = index + 1;
    const result: Order = await service.get({ id });
    const request = requests[index];
    const expectedUrl = `https://example.invalid/invoke/example.versions.OrderApiService${version}/get`;
    if (request?.url !== expectedUrl) {
      throw new Error(`Expected ${expectedUrl}, received ${request?.url}`);
    }
    if (request.method !== 'POST' || request.body !== JSON.stringify({ params: { id } })) {
      throw new Error(`Incorrect request for ${expectedUrl}`);
    }
    if (result.id !== id) {
      throw new Error(`Incorrect response for ${expectedUrl}: ${JSON.stringify(result)}`);
    }
  }
  if (requests.length !== services.length) {
    throw new Error(`Expected ${services.length} calls, received ${requests.length}`);
  }
}
