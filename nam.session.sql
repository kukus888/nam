select h.id as hc_id, s.hostname as hostname, ad.port as port, h.url as url from healthcheck h 
left join application_definition ad  on ad.healthcheck_id = h.id
left join application_instance ai on application_definition_id = ad.id
left join "server" s on s.id = ai.server_id
where h.id = 2