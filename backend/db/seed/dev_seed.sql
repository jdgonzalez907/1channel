-- ============================================================================
-- 1Channel - Seed de datos de prueba
--
-- SOLO DESARROLLO. DESTRUCTIVO: trunca users/contacts/conversations/messages/
-- personal_information. NUNCA es una migracion y NUNCA debe correr en produccion.
--
-- Uso:  make seed
--
-- Genera: 3 agentes, 15 personas (5 compartidas por 2 contactos), 50 contactos
-- (con y sin display_name), 90 conversaciones y 3000 mensajes
-- (5..50 por conversacion), con todos los invariantes de dominio respetados.
-- IDs y external_ids en UUID v7 (uuidv7(), nativo de Postgres 18).
-- ============================================================================

BEGIN;

-- Guarda: abortar si la base no parece de desarrollo ------------------------
DO $$
BEGIN
    IF current_database() NOT LIKE '%dev%' THEN
        RAISE EXCEPTION 'seed abortado: la base "%" no parece de desarrollo', current_database();
    END IF;
END $$;

TRUNCATE users, contacts, conversations, messages, personal_information RESTART IDENTITY CASCADE;

-- 1. Agentes, contactos y catalogo de frases --------------------------------
CREATE TEMP TABLE seed_user AS
SELECT g                                          AS agent_no,
       uuidv7()                                   AS id,
       now() - interval '200 days' + (g * interval '1 hour') AS created_at
FROM generate_series(1, 3) g;

INSERT INTO users (id, created_at)
SELECT id, created_at FROM seed_user;

-- Personas: 15 en total. Las 5 primeras las comparten pares de contactos
-- (1..10); las 10 restantes pertenecen a un contacto cada una (11..20).
CREATE TEMP TABLE seed_pi AS
SELECT p AS pi_no,
       uuidv7() AS id,
       lpad((10000000 + p)::text, 8, '0') AS identification_number,
       (ARRAY['Juan','Maria','Carlos','Lucia','Pedro','Ana','Diego','Sofia','Jorge','Valentina','Andres','Camila','Mateo','Isabel','Ricardo'])[p] AS first_name,
       (ARRAY['Perez','Gomez','Lopez','Martinez','Rodriguez','Fernandez','Garcia','Diaz','Torres','Ramirez','Vargas','Castro','Rojas','Silva','Mora'])[p] AS last_name,
       CASE WHEN p % 5 = 0 THEN NULL ELSE '555-' || lpad((1000 + p)::text, 4, '0') END AS phone_number,
       CASE WHEN p % 4 = 0 THEN NULL ELSE 'persona' || p || '@example.com' END AS email,
       CASE WHEN p % 6 = 0 THEN NULL ELSE 'Calle ' || p || ' #' || (10 + p) || '-' || (20 + p) END AS address
FROM generate_series(1, 15) p;

INSERT INTO personal_information (id, identification_number, first_name, last_name,
                                  phone_number, email, address, created_at, updated_at)
SELECT id, identification_number, first_name, last_name, phone_number, email, address,
       now() - interval '180 days' + (pi_no * interval '1 hour'),
       now() - interval '30 days' + (pi_no * interval '1 hour')
FROM seed_pi;

-- Contactos 1..20 tienen persona (1..10 compartida en pares, 11..20 propia);
-- 21..44 solo display_name; 45..50 (la minoria) sin display_name ni persona.
CREATE TEMP TABLE seed_contact AS
SELECT g                                          AS contact_no,
       uuidv7()                                   AS id,
       uuidv7()::text                             AS external_contact_id,
       CASE WHEN g <= 20 OR g >= 45 THEN NULL
            ELSE (ARRAY['Cliente','Usuario','Contacto'])[1 + (g % 3)] || ' ' || lpad((1000 + g)::text, 4, '0')
       END                                        AS display_name,
       CASE WHEN g <= 10 THEN (g + 1) / 2
            WHEN g <= 20 THEN 5 + (g - 10)
            ELSE NULL
       END                                        AS pi_no,
       now() - interval '200 days' + (g * interval '1 hour') AS created_at
FROM generate_series(1, 50) g;

INSERT INTO contacts (id, external_contact_id, display_name, personal_information_id, created_at)
SELECT sc.id, sc.external_contact_id, sc.display_name, pi.id, sc.created_at
FROM seed_contact sc
LEFT JOIN seed_pi pi ON pi.pi_no = sc.pi_no;

CREATE TEMP TABLE seed_phrase (topic text, role text, idx int, body text);

INSERT INTO seed_phrase (topic, role, idx, body) VALUES
  ('stock', 'contact', 1, 'Hola, tienen la PS5 en stock?'),
  ('stock', 'contact', 2, 'Buenas, les queda la Xbox Series X?'),
  ('stock', 'contact', 3, 'Tienen disponible la Nintendo Switch OLED?'),
  ('stock', 'contact', 4, 'Hay stock de la PS5 Slim digital?'),
  ('stock', 'contact', 5, 'Ya llego la Xbox Series S?'),
  ('stock', 'agent',   1, 'Hola, con gusto te ayudo. Si, tenemos unidades disponibles.'),
  ('stock', 'agent',   2, 'Claro, dejame revisar el inventario y te confirmo.'),
  ('stock', 'agent',   3, 'Si, hay stock. Prefieres la version con lector o digital?'),
  ('stock', 'agent',   4, 'Tenemos disponibilidad para entrega inmediata.'),

  ('precio', 'contact', 1, 'Cuanto cuesta la PS5?'),
  ('precio', 'contact', 2, 'La Switch OLED esta en oferta?'),
  ('precio', 'contact', 3, 'Me pueden dar precio de contado?'),
  ('precio', 'contact', 4, 'Tienen algun descuento por pago en efectivo?'),
  ('precio', 'contact', 5, 'El precio incluye el control extra?'),
  ('precio', 'agent',   1, 'Claro, el precio actual es con envio incluido.'),
  ('precio', 'agent',   2, 'Si, tenemos una promo vigente hasta fin de mes.'),
  ('precio', 'agent',   3, 'Te puedo ofrecer un descuento por pago en efectivo.'),
  ('precio', 'agent',   4, 'Dejame confirmarte el precio con impuestos.'),

  ('envio', 'contact', 1, 'Cuanto tarda el envio a Medellin?'),
  ('envio', 'contact', 2, 'Hacen envios a todo el pais?'),
  ('envio', 'contact', 3, 'Puedo recoger en tienda?'),
  ('envio', 'contact', 4, 'El envio es gratis?'),
  ('envio', 'contact', 5, 'Como va mi pedido?'),
  ('envio', 'agent',   1, 'El envio a Medellin tarda de 2 a 3 dias habiles.'),
  ('envio', 'agent',   2, 'Si, enviamos a todo el pais.'),
  ('envio', 'agent',   3, 'Puedes recoger en tienda sin costo adicional.'),
  ('envio', 'agent',   4, 'Tu pedido ya esta en camino, te comparto la guia.'),

  ('pago', 'contact', 1, 'Puedo pagar a cuotas?'),
  ('pago', 'contact', 2, 'Aceptan tarjeta de credito?'),
  ('pago', 'contact', 3, 'Se puede pagar contra entrega?'),
  ('pago', 'contact', 4, 'Reciben transferencia?'),
  ('pago', 'contact', 5, 'Cual es el minimo de la cuota inicial?'),
  ('pago', 'agent',   1, 'Si, aceptamos tarjeta y pago a cuotas.'),
  ('pago', 'agent',   2, 'Puedes pagar contra entrega o por transferencia.'),
  ('pago', 'agent',   3, 'Si, manejamos hasta 12 cuotas sin interes.'),
  ('pago', 'agent',   4, 'Con gusto te explico las opciones de pago.'),

  ('garantia', 'contact', 1, 'La consola viene con garantia?'),
  ('garantia', 'contact', 2, 'Se rayo el lente del lector, lo cubre la garantia?'),
  ('garantia', 'contact', 3, 'Cuanto tiempo de garantia tiene?'),
  ('garantia', 'contact', 4, 'La consola no enciende, la puedo llevar?'),
  ('garantia', 'contact', 5, 'La garantia cubre el control?'),
  ('garantia', 'agent',   1, 'Si, todas las consolas tienen 12 meses de garantia.'),
  ('garantia', 'agent',   2, 'Claro, la garantia cubre fallas de fabrica.'),
  ('garantia', 'agent',   3, 'Puedes acercarte a la tienda con la factura.'),
  ('garantia', 'agent',   4, 'Vamos a revisarlo, si es de fabrica lo cubrimos.'),

  ('devolucion', 'contact', 1, 'Quiero devolver la consola.'),
  ('devolucion', 'contact', 2, 'Llego con una falla, quiero el reembolso.'),
  ('devolucion', 'contact', 3, 'Puedo cambiarla por otro modelo?'),
  ('devolucion', 'contact', 4, 'No era lo que esperaba, aceptan devolucion?'),
  ('devolucion', 'contact', 5, 'Cual es el plazo para devolver?'),
  ('devolucion', 'agent',   1, 'Lamento el inconveniente, gestionemos la devolucion.'),
  ('devolucion', 'agent',   2, 'Si, tienes 30 dias para devoluciones.'),
  ('devolucion', 'agent',   3, 'Te ayudo con el cambio por otro modelo.'),
  ('devolucion', 'agent',   4, 'Procesamos el reembolso al recibir el producto.'),

  ('accesorios', 'contact', 1, 'Tienen control extra para la PS5?'),
  ('accesorios', 'contact', 2, 'Venden juegos para Xbox?'),
  ('accesorios', 'contact', 3, 'Tienen base de carga para los mandos?'),
  ('accesorios', 'contact', 4, 'El paquete incluye audifonos?'),
  ('accesorios', 'contact', 5, 'Tienen cables HDMI adicionales?'),
  ('accesorios', 'agent',   1, 'Si, tenemos controles y accesorios disponibles.'),
  ('accesorios', 'agent',   2, 'Claro, tambien manejamos juegos y mandos.'),
  ('accesorios', 'agent',   3, 'Te recomiendo el paquete que incluye base de carga.'),
  ('accesorios', 'agent',   4, 'Si, el combo trae audifonos de regalo.');

-- 2. Plan de conversaciones -------------------------------------------------
-- 90 conversaciones:
--   conv 1..30   finished de los contactos 1..30
--   conv 31..50  finished de los contactos 31..50
--   conv 51..60  finished (segunda) de los contactos 31..40
--   conv 61..72  pending de los contactos 1..12
--   conv 73..90  assigned de los contactos 13..30
CREATE TEMP TABLE seed_conv (
    conv_no    int PRIMARY KEY,
    contact_no int,
    status     text,
    agent_no   int,
    topic      text,
    created_at timestamptz,
    finished_at timestamptz,
    updated_at timestamptz,
    msg_count  int,
    id         uuid
);

INSERT INTO seed_conv (conv_no, contact_no, status, topic)
SELECT g,
       CASE WHEN g <= 50 THEN g
            WHEN g <= 60 THEN 31 + (g - 51)
            ELSE g - 60 END,
       CASE WHEN g BETWEEN 1 AND 10 THEN 'resolved'
            WHEN g BETWEEN 11 AND 30 THEN 'expired'
            WHEN g BETWEEN 31 AND 50 THEN 'resolved'
            WHEN g BETWEEN 51 AND 60 THEN 'expired'
            WHEN g BETWEEN 61 AND 72 THEN 'pending'
            ELSE 'assigned' END,
       (ARRAY['stock','precio','envio','pago','garantia','devolucion','accesorios'])[1 + (g % 7)]
FROM generate_series(1, 90) g;

UPDATE seed_conv
SET created_at = now() - interval '150 days'
               + (conv_no * interval '1 day')
               + (interval '1 hour' * (conv_no % 7));

-- Agentes: assigned 6/6/6, resolved 10/10/10, expired-con-agente 7/7/6.
-- Las expired sin agente (conv 21..30) quedan con agent_no NULL.
UPDATE seed_conv sc
SET agent_no = sub.rn
FROM (
    SELECT conv_no,
           1 + ((row_number() OVER (PARTITION BY status ORDER BY conv_no) - 1) % 3) AS rn
    FROM seed_conv
    WHERE status IN ('assigned', 'resolved', 'expired')
      AND NOT (status = 'expired' AND conv_no BETWEEN 21 AND 30)
) sub
WHERE sc.conv_no = sub.conv_no;

-- 3. Mensajes ---------------------------------------------------------------
CREATE TEMP TABLE seed_msg (
    conv_no      int,
    m_idx        int,
    owner        text,
    sent_at      timestamptz,
    external_id  text,
    body         text,
    flag_failed  boolean,
    flag_deleted boolean,
    flag_edited  boolean,
    read_at      timestamptz,
    status       text,
    deleted_at   timestamptz,
    edited_at    timestamptz,
    id           uuid
);

DO $$
DECLARE
    c            record;
    v_n          int;
    v_has_agent  boolean;
    v_owners     text[];
    v_owner      text;
    v_prev       text;
    v_run        int;
    v_i          int;
    j            int;
    v_t          timestamptz;
    v_ext        text;
    v_body       text;
    v_failed     boolean;
    v_deleted    boolean;
    v_edited     boolean;
    v_total      int;
    v_no         int;
    v_role_count int;
BEGIN
    -- 3a. conteo de mensajes por conversacion (dentro de las bandas)
    FOR c IN SELECT conv_no, status FROM seed_conv ORDER BY conv_no LOOP
        CASE c.status
            WHEN 'pending'  THEN v_n := 5  + floor(random() * 6)::int;   -- 5..10
            WHEN 'assigned' THEN v_n := 15 + floor(random() * 31)::int;  -- 15..45
            ELSE                 v_n := 20 + floor(random() * 31)::int;  -- 20..50
        END CASE;
        UPDATE seed_conv SET msg_count = v_n WHERE conv_no = c.conv_no;
    END LOOP;

    SELECT sum(msg_count) INTO v_total FROM seed_conv;

    -- 3b. ajustar la suma exacta a 3000 respetando [5, 50]
    FOR v_i IN 1..200000 LOOP
        EXIT WHEN v_total = 3000;
        IF v_total < 3000 THEN
            SELECT conv_no INTO v_no FROM seed_conv WHERE msg_count < 50 ORDER BY random() LIMIT 1;
            IF v_no IS NULL THEN RAISE EXCEPTION 'no se pudo ajustar a 3000 (tope alcanzado)'; END IF;
            UPDATE seed_conv SET msg_count = msg_count + 1 WHERE conv_no = v_no;
            v_total := v_total + 1;
        ELSE
            SELECT conv_no INTO v_no FROM seed_conv WHERE msg_count > 5 ORDER BY random() LIMIT 1;
            IF v_no IS NULL THEN RAISE EXCEPTION 'no se pudo ajustar a 3000 (minimo alcanzado)'; END IF;
            UPDATE seed_conv SET msg_count = msg_count - 1 WHERE conv_no = v_no;
            v_total := v_total - 1;
        END IF;
    END LOOP;
    IF v_total <> 3000 THEN RAISE EXCEPTION 'total de mensajes % distinto de 3000', v_total; END IF;

    -- 3c. generar mensajes
    FOR c IN
        SELECT conv_no, status, agent_no, topic, created_at, msg_count
        FROM seed_conv ORDER BY conv_no
    LOOP
        v_n := c.msg_count;
        v_has_agent := c.agent_no IS NOT NULL;

        -- duenos: el contacto inicia; con agente alternan runs de 1..2
        v_owners := ARRAY[]::text[];
        v_prev := NULL;
        v_i := 1;
        WHILE v_i <= v_n LOOP
            IF NOT v_has_agent THEN
                v_owner := 'contact';
                v_run := v_n - v_i + 1;
            ELSIF v_i = 1 THEN
                v_owner := 'contact';
                v_run := 1 + floor(random() * 2)::int;
            ELSE
                v_owner := CASE WHEN v_prev = 'contact' THEN 'agent' ELSE 'contact' END;
                v_run := 1 + floor(random() * 2)::int;
            END IF;

            IF v_i + v_run - 1 > v_n THEN
                v_run := v_n - v_i + 1;
            END IF;

            FOR j IN v_i .. (v_i + v_run - 1) LOOP
                v_owners[j] := v_owner;
            END LOOP;

            v_prev := v_owner;
            v_i := v_i + v_run;
        END LOOP;

        -- timestamps crecientes
        v_t := c.created_at + (interval '1 minute' * (10 + floor(random() * 50)::int));
        FOR j IN 1..v_n LOOP
            IF j > 1 THEN
                v_t := v_t + (interval '1 minute' * (5 + floor(random() * 175)::int));
            END IF;

            v_owner := v_owners[j];
            v_failed := false;
            v_deleted := false;
            v_edited := false;

            -- trazas: failed solo de agente; deletes/edits de agente solo en open (assigned)
            IF v_owner = 'agent' AND ((c.conv_no + j) % 43 = 0) THEN
                v_failed := true;
            ELSIF v_owner = 'agent' AND c.status <> 'assigned' THEN
                v_failed := false;
            ELSIF ((c.conv_no + j) % 37 = 0) THEN
                v_deleted := true;
            ELSIF ((c.conv_no + j) % 41 = 0) THEN
                v_edited := true;
            END IF;

            IF v_owner = 'contact' THEN
                v_ext := uuidv7()::text;
            ELSIF v_failed THEN
                v_ext := NULL;
            ELSE
                v_ext := uuidv7()::text;
            END IF;

            SELECT count(*) INTO v_role_count
            FROM seed_phrase WHERE topic = c.topic AND role = v_owner;

            SELECT p.body INTO v_body
            FROM seed_phrase p
            WHERE p.topic = c.topic AND p.role = v_owner
            ORDER BY p.idx
            OFFSET ((j - 1) % v_role_count)
            LIMIT 1;

            INSERT INTO seed_msg (conv_no, m_idx, owner, sent_at, external_id, body,
                                  flag_failed, flag_deleted, flag_edited)
            VALUES (c.conv_no, j, v_owner, v_t, v_ext, v_body, v_failed, v_deleted, v_edited);
        END LOOP;
    END LOOP;
END $$;

-- 4. Lectura, estado y trazas ----------------------------------------------
-- read_at: cada mensaje del contacto se lee en el primer mensaje de agente posterior.
UPDATE seed_msg m
SET read_at = (
    SELECT MIN(a.sent_at)
    FROM seed_msg a
    WHERE a.conv_no = m.conv_no AND a.m_idx > m.m_idx AND a.owner = 'agent'
)
WHERE m.owner = 'contact'
  AND EXISTS (
      SELECT 1 FROM seed_conv c WHERE c.conv_no = m.conv_no AND c.agent_no IS NOT NULL
  );

-- read_at del agente: cada mensaje del agente se lee en el primer mensaje del
-- contacto posterior (el contacto lo vio).
UPDATE seed_msg m
SET read_at = (
    SELECT MIN(a.sent_at)
    FROM seed_msg a
    WHERE a.conv_no = m.conv_no AND a.m_idx > m.m_idx AND a.owner = 'contact'
)
WHERE m.owner = 'agent'
  AND NOT m.flag_failed
  AND NOT m.flag_deleted;

UPDATE seed_msg
SET status = CASE
    WHEN flag_failed  THEN 'failed'
    WHEN flag_deleted THEN 'deleted'
    WHEN read_at IS NOT NULL THEN 'read'
    ELSE 'sent'
END;

UPDATE seed_msg
SET deleted_at = sent_at + (interval '1 minute' * (30 + floor(random() * 600)::int))
WHERE flag_deleted;

UPDATE seed_msg
SET edited_at = sent_at + (interval '1 minute' * (30 + floor(random() * 600)::int))
WHERE flag_edited;

UPDATE seed_msg  SET id = uuidv7();
UPDATE seed_conv SET id = uuidv7();

-- 5. finished_at y updated_at ----------------------------------------------
UPDATE seed_conv c
SET finished_at = (SELECT max(sent_at) FROM seed_msg m WHERE m.conv_no = c.conv_no) + interval '3 hours'
WHERE c.status IN ('resolved', 'expired');

UPDATE seed_conv c
SET updated_at = GREATEST(
    (SELECT max(sent_at)   FROM seed_msg m WHERE m.conv_no = c.conv_no),
    c.finished_at,
    (SELECT max(read_at)   FROM seed_msg m WHERE m.conv_no = c.conv_no),
    (SELECT max(edited_at) FROM seed_msg m WHERE m.conv_no = c.conv_no),
    (SELECT max(deleted_at) FROM seed_msg m WHERE m.conv_no = c.conv_no)
);

-- 6. Persistir conversaciones y mensajes ------------------------------------
INSERT INTO conversations (id, status, user_id, contact_id, created_at, updated_at, finished_at)
SELECT c.id, c.status, u.id, ct.id, c.created_at, c.updated_at, c.finished_at
FROM seed_conv c
JOIN seed_contact ct ON ct.contact_no = c.contact_no
LEFT JOIN seed_user u ON u.agent_no = c.agent_no
ORDER BY c.conv_no;

INSERT INTO messages (id, conversation_id, status, type, text, user_id, contact_id,
                      external_id, sent_at, read_at, edited_at, deleted_at)
SELECT m.id, c.id, m.status, 'text', m.body,
       CASE WHEN m.owner = 'agent'   THEN u.id  END,
       CASE WHEN m.owner = 'contact' THEN ct.id END,
       m.external_id, m.sent_at, m.read_at, m.edited_at, m.deleted_at
FROM seed_msg m
JOIN seed_conv c ON c.conv_no = m.conv_no
JOIN seed_contact ct ON ct.contact_no = c.contact_no
LEFT JOIN seed_user u ON u.agent_no = c.agent_no
ORDER BY m.conv_no, m.m_idx;

-- 7. Modelo de lectura denormalizado (= RefreshConversationLastMessage) -----
UPDATE conversations c
SET last_message_id = lm.id,
    last_message_at = lm.sent_at,
    unread_count    = COALESCE(u.cnt, 0)
FROM (
    SELECT DISTINCT ON (conversation_id) conversation_id, id, sent_at
    FROM messages
    ORDER BY conversation_id, sent_at DESC, id DESC
) lm
LEFT JOIN (
    SELECT conversation_id, count(*) AS cnt
    FROM messages
    WHERE contact_id IS NOT NULL AND read_at IS NULL
    GROUP BY conversation_id
) u ON u.conversation_id = lm.conversation_id
WHERE c.id = lm.conversation_id;

-- 8. Validaciones -----------------------------------------------------------
DO $$
DECLARE
    v_count int;
    v_bad   int;
BEGIN
    SELECT count(*) INTO v_count FROM users;
    IF v_count <> 3 THEN RAISE EXCEPTION 'usuarios = % (esperado 3)', v_count; END IF;

    SELECT count(*) INTO v_count FROM contacts;
    IF v_count <> 50 THEN RAISE EXCEPTION 'contactos = % (esperado 50)', v_count; END IF;

    SELECT count(*) INTO v_count FROM personal_information;
    IF v_count <> 15 THEN RAISE EXCEPTION 'personal information = % (esperado 15)', v_count; END IF;

    SELECT count(*) INTO v_count FROM contacts WHERE personal_information_id IS NOT NULL;
    IF v_count <> 20 THEN RAISE EXCEPTION 'contactos con persona = % (esperado 20)', v_count; END IF;

    SELECT count(*) INTO v_bad FROM (
        SELECT personal_information_id FROM contacts
        WHERE personal_information_id IS NOT NULL
        GROUP BY personal_information_id HAVING count(*) > 1
    ) x;
    IF v_bad <> 5 THEN RAISE EXCEPTION 'personas compartidas = % (esperado 5)', v_bad; END IF;

    SELECT count(*) INTO v_count FROM conversations;
    IF v_count <> 90 THEN RAISE EXCEPTION 'conversaciones = % (esperado 90)', v_count; END IF;

    SELECT count(*) INTO v_count FROM messages;
    IF v_count <> 3000 THEN RAISE EXCEPTION 'mensajes = % (esperado 3000)', v_count; END IF;

    SELECT count(*) INTO v_bad FROM (
        SELECT conversation_id FROM messages
        GROUP BY conversation_id HAVING count(*) < 5 OR count(*) > 50
    ) x;
    IF v_bad > 0 THEN RAISE EXCEPTION 'conversaciones fuera de [5,50]: %', v_bad; END IF;

    SELECT count(*) INTO v_bad FROM (
        SELECT contact_id FROM conversations
        WHERE status IN ('pending', 'assigned') AND contact_id IS NOT NULL
        GROUP BY contact_id HAVING count(*) > 1
    ) x;
    IF v_bad > 0 THEN RAISE EXCEPTION 'contactos con mas de una conversacion abierta: %', v_bad; END IF;

    SELECT count(*) INTO v_bad
    FROM conversations c
    WHERE c.last_message_id IS DISTINCT FROM (
              SELECT m.id FROM messages m WHERE m.conversation_id = c.id
              ORDER BY m.sent_at DESC, m.id DESC LIMIT 1)
       OR c.last_message_at IS DISTINCT FROM (
              SELECT m.sent_at FROM messages m WHERE m.conversation_id = c.id
              ORDER BY m.sent_at DESC, m.id DESC LIMIT 1)
       OR c.unread_count IS DISTINCT FROM (
              SELECT count(*) FROM messages m
              WHERE m.conversation_id = c.id AND m.contact_id IS NOT NULL AND m.read_at IS NULL);
    IF v_bad > 0 THEN RAISE EXCEPTION 'denormalizado inconsistente en % conversaciones', v_bad; END IF;

    SELECT count(*) INTO v_bad
    FROM messages m JOIN conversations c ON c.id = m.conversation_id
    WHERE c.finished_at IS NOT NULL AND m.sent_at >= c.finished_at;
    IF v_bad > 0 THEN RAISE EXCEPTION 'mensajes no anteriores a finished_at: %', v_bad; END IF;

    SELECT count(*) INTO v_bad FROM messages WHERE read_at IS NOT NULL AND read_at < sent_at;
    IF v_bad > 0 THEN RAISE EXCEPTION 'read_at anterior a sent_at: %', v_bad; END IF;

    SELECT count(*) INTO v_bad
    FROM messages WHERE status = 'failed' AND (user_id IS NULL OR external_id IS NOT NULL);
    IF v_bad > 0 THEN RAISE EXCEPTION 'mensajes failed invalidos: %', v_bad; END IF;

    SELECT count(*) INTO v_bad FROM messages WHERE (user_id IS NULL) = (contact_id IS NULL);
    IF v_bad > 0 THEN RAISE EXCEPTION 'mensajes sin dueno unico: %', v_bad; END IF;

    SELECT count(*) INTO v_bad
    FROM conversations
    WHERE (status IN ('resolved', 'expired') AND finished_at IS NULL)
       OR (status IN ('pending', 'assigned') AND finished_at IS NOT NULL);
    IF v_bad > 0 THEN RAISE EXCEPTION 'finished_at inconsistente: %', v_bad; END IF;

    SELECT count(*) INTO v_bad FROM conversations WHERE finished_at IS NOT NULL AND updated_at < finished_at;
    IF v_bad > 0 THEN RAISE EXCEPTION 'updated_at anterior a finished_at: %', v_bad; END IF;
END $$;

COMMIT;
