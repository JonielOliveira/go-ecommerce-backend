-- ============================================================
-- SEED DE DESENVOLVIMENTO
--
-- Cria:
--   - 10 usuários: 5 admins e 5 customers
--   - 20 produtos de supermercado
--   - 10 pedidos
--   - 3 a 8 itens por pedido
--
-- Senha de todos os usuários:
--   Senha@123
--
-- O hash abaixo foi gerado com bcrypt cost 12.
--
-- ATENÇÃO:
--   Este arquivo é destinado apenas ao ambiente de
--   desenvolvimento/testes.
--
--   Ao ser executado novamente:
--   - os pedidos deste seed são apagados e recriados;
--   - os estoques dos produtos deste seed são redefinidos;
--   - pedidos CANCELED não reduzem o estoque atual;
--   - pedidos PENDING e PAID reduzem o estoque atual.
-- ============================================================

BEGIN;

-- ============================================================
-- 1. REMOVER PEDIDOS ANTERIORES DESTE SEED
--
-- order_items serão removidos automaticamente por ON DELETE
-- CASCADE.
-- ============================================================

DELETE FROM orders
WHERE id IN (
    '20000000-0000-7000-8000-000000000001',
    '20000000-0000-7000-8000-000000000002',
    '20000000-0000-7000-8000-000000000003',
    '20000000-0000-7000-8000-000000000004',
    '20000000-0000-7000-8000-000000000005',
    '20000000-0000-7000-8000-000000000006',
    '20000000-0000-7000-8000-000000000007',
    '20000000-0000-7000-8000-000000000008',
    '20000000-0000-7000-8000-000000000009',
    '20000000-0000-7000-8000-000000000010'
);

-- ============================================================
-- 2. USUÁRIOS
-- ============================================================

INSERT INTO users (
    id,
    name,
    email,
    email_verified_at,
    avatar_url,
    role,
    active,
    last_login_at,
    created_at,
    updated_at,
    deleted_at
)
VALUES
    (
        '00000000-0000-7000-8000-000000000001',
        'Ana Souza',
        'ana.admin@gmail.com',
        NOW(),
        NULL,
        'admin',
        TRUE,
        NULL,
        NOW() - INTERVAL '90 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000002',
        'Bruno Lima',
        'bruno.admin@gmail.com',
        NOW(),
        NULL,
        'admin',
        TRUE,
        NULL,
        NOW() - INTERVAL '85 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000003',
        'Carla Mendes',
        'carla.admin@gmail.com',
        NOW(),
        NULL,
        'admin',
        TRUE,
        NULL,
        NOW() - INTERVAL '80 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000004',
        'Diego Rocha',
        'diego.admin@gmail.com',
        NOW(),
        NULL,
        'admin',
        TRUE,
        NULL,
        NOW() - INTERVAL '75 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000005',
        'Elisa Martins',
        'elisa.admin@gmail.com',
        NOW(),
        NULL,
        'admin',
        TRUE,
        NULL,
        NOW() - INTERVAL '70 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000006',
        'Fernanda Alves',
        'fernanda.customer@gmail.com',
        NOW(),
        NULL,
        'customer',
        TRUE,
        NULL,
        NOW() - INTERVAL '65 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000007',
        'Gustavo Nunes',
        'gustavo.customer@gmail.com',
        NOW(),
        NULL,
        'customer',
        TRUE,
        NULL,
        NOW() - INTERVAL '60 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000008',
        'Helena Costa',
        'helena.customer@gmail.com',
        NOW(),
        NULL,
        'customer',
        TRUE,
        NULL,
        NOW() - INTERVAL '55 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000009',
        'Igor Ribeiro',
        'igor.customer@gmail.com',
        NOW(),
        NULL,
        'customer',
        TRUE,
        NULL,
        NOW() - INTERVAL '50 days',
        NOW(),
        NULL
    ),
    (
        '00000000-0000-7000-8000-000000000010',
        'Juliana Freitas',
        'juliana.customer@gmail.com',
        NOW(),
        NULL,
        'customer',
        TRUE,
        NULL,
        NOW() - INTERVAL '45 days',
        NOW(),
        NULL
    )
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    email_verified_at = EXCLUDED.email_verified_at,
    avatar_url = EXCLUDED.avatar_url,
    role = EXCLUDED.role,
    active = EXCLUDED.active,
    deleted_at = NULL,
    updated_at = NOW();

-- ============================================================
-- 3. CREDENCIAIS
--
-- Senha de todos:
--   Senha@123
--
-- Hash bcrypt cost 12:
--   $2y$12$rCB0nxbDiqEdZSNP844CDu1dKHwCGAr6.188ZEgi2RCekvTXp1PWO
-- ============================================================

INSERT INTO user_password_credentials (
    user_id,
    password_hash,
    password_changed_at
)
SELECT
    id,
    '$2y$12$rCB0nxbDiqEdZSNP844CDu1dKHwCGAr6.188ZEgi2RCekvTXp1PWO',
    NOW()
FROM users
WHERE id IN (
    '00000000-0000-7000-8000-000000000001',
    '00000000-0000-7000-8000-000000000002',
    '00000000-0000-7000-8000-000000000003',
    '00000000-0000-7000-8000-000000000004',
    '00000000-0000-7000-8000-000000000005',
    '00000000-0000-7000-8000-000000000006',
    '00000000-0000-7000-8000-000000000007',
    '00000000-0000-7000-8000-000000000008',
    '00000000-0000-7000-8000-000000000009',
    '00000000-0000-7000-8000-000000000010'
)
ON CONFLICT (user_id) DO UPDATE
SET
    password_hash = EXCLUDED.password_hash,
    password_changed_at = EXCLUDED.password_changed_at;

-- ============================================================
-- 4. PRODUTOS
--
-- Os estoques informados aqui representam o estoque-base antes
-- da criação dos pedidos do seed.
-- ============================================================

INSERT INTO products (
    id,
    name,
    description,
    price,
    stock,
    category_id,
    image_url,
    active,
    created_at,
    updated_at,
    deleted_at
)
VALUES
    (
        '10000000-0000-7000-8000-000000000001',
        'Arroz tipo 1 5 kg',
        'Arroz branco tipo 1, pacote com 5 kg.',
        29.90,
        80,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '120 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000002',
        'Feijão carioca 1 kg',
        'Feijão carioca selecionado, pacote com 1 kg.',
        8.49,
        100,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '118 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000003',
        'Leite em pó integral 400 g',
        'Leite em pó integral instantâneo, embalagem de 400 g.',
        17.90,
        60,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '116 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000004',
        'Leite integral UHT 1 L',
        'Leite integral longa vida, embalagem de 1 litro.',
        5.49,
        120,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '114 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000005',
        'Açúcar refinado 1 kg',
        'Açúcar refinado especial, pacote com 1 kg.',
        4.79,
        90,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '112 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000006',
        'Café torrado e moído 500 g',
        'Café torrado e moído tradicional, pacote com 500 g.',
        18.90,
        70,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '110 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000007',
        'Óleo de soja 900 ml',
        'Óleo de soja refinado, garrafa com 900 ml.',
        7.29,
        90,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '108 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000008',
        'Macarrão espaguete 500 g',
        'Macarrão de sêmola tipo espaguete, pacote com 500 g.',
        4.99,
        100,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '106 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000009',
        'Molho de tomate 300 g',
        'Molho de tomate tradicional, sachê com 300 g.',
        2.79,
        110,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '104 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000010',
        'Ervilha em conserva 170 g',
        'Ervilha em conserva, lata com peso drenado de 170 g.',
        3.89,
        80,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '102 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000011',
        'Milho verde em conserva 170 g',
        'Milho verde em conserva, lata com peso drenado de 170 g.',
        3.79,
        80,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '100 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000012',
        'Sardinha em óleo 125 g',
        'Sardinha em óleo comestível, lata com 125 g.',
        6.99,
        60,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '98 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000013',
        'Farinha de trigo 1 kg',
        'Farinha de trigo tradicional, pacote com 1 kg.',
        5.69,
        75,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '96 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000014',
        'Sal refinado 1 kg',
        'Sal refinado iodado, pacote com 1 kg.',
        2.49,
        100,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '94 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000015',
        'Biscoito cream cracker 350 g',
        'Biscoito salgado tipo cream cracker, pacote com 350 g.',
        5.99,
        85,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '92 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000016',
        'Achocolatado em pó 400 g',
        'Achocolatado em pó instantâneo, embalagem com 400 g.',
        8.99,
        65,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '90 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000017',
        'Aveia em flocos 500 g',
        'Aveia integral em flocos, embalagem com 500 g.',
        7.49,
        70,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '88 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000018',
        'Atum sólido em óleo 170 g',
        'Atum sólido em óleo, lata com 170 g.',
        9.89,
        55,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '86 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000019',
        'Farofa pronta temperada 500 g',
        'Farofa de mandioca pronta e temperada, pacote com 500 g.',
        7.99,
        60,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '84 days',
        NOW(),
        NULL
    ),
    (
        '10000000-0000-7000-8000-000000000020',
        'Leite condensado 395 g',
        'Leite condensado integral, embalagem com 395 g.',
        6.49,
        75,
        NULL,
        NULL,
        TRUE,
        NOW() - INTERVAL '82 days',
        NOW(),
        NULL
    )
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,

    -- Redefine o estoque-base para tornar o seed reexecutável.
    stock = EXCLUDED.stock,

    category_id = EXCLUDED.category_id,
    image_url = EXCLUDED.image_url,
    active = EXCLUDED.active,
    deleted_at = NULL,
    updated_at = NOW();

-- ============================================================
-- 5. PEDIDOS
--
-- Um pedido para cada usuário.
--
-- Distribuição:
--   4 PENDING
--   4 PAID
--   2 CANCELED
--
-- total_amount começa em 0 e será calculado a partir dos itens.
-- ============================================================

INSERT INTO orders (
    id,
    customer_id,
    status,
    total_amount,
    paid_at,
    canceled_at,
    created_at,
    updated_at
)
VALUES
    (
        '20000000-0000-7000-8000-000000000001',
        '00000000-0000-7000-8000-000000000001',
        'PENDING',
        0,
        NULL,
        NULL,
        NOW() - INTERVAL '10 days',
        NOW() - INTERVAL '10 days'
    ),
    (
        '20000000-0000-7000-8000-000000000002',
        '00000000-0000-7000-8000-000000000002',
        'PAID',
        0,
        NOW() - INTERVAL '8 days 20 hours',
        NULL,
        NOW() - INTERVAL '9 days',
        NOW() - INTERVAL '8 days 20 hours'
    ),
    (
        '20000000-0000-7000-8000-000000000003',
        '00000000-0000-7000-8000-000000000003',
        'CANCELED',
        0,
        NULL,
        NOW() - INTERVAL '7 days 18 hours',
        NOW() - INTERVAL '8 days',
        NOW() - INTERVAL '7 days 18 hours'
    ),
    (
        '20000000-0000-7000-8000-000000000004',
        '00000000-0000-7000-8000-000000000004',
        'PENDING',
        0,
        NULL,
        NULL,
        NOW() - INTERVAL '7 days',
        NOW() - INTERVAL '7 days'
    ),
    (
        '20000000-0000-7000-8000-000000000005',
        '00000000-0000-7000-8000-000000000005',
        'PAID',
        0,
        NOW() - INTERVAL '5 days 12 hours',
        NULL,
        NOW() - INTERVAL '6 days',
        NOW() - INTERVAL '5 days 12 hours'
    ),
    (
        '20000000-0000-7000-8000-000000000006',
        '00000000-0000-7000-8000-000000000006',
        'CANCELED',
        0,
        NULL,
        NOW() - INTERVAL '4 days 18 hours',
        NOW() - INTERVAL '5 days',
        NOW() - INTERVAL '4 days 18 hours'
    ),
    (
        '20000000-0000-7000-8000-000000000007',
        '00000000-0000-7000-8000-000000000007',
        'PENDING',
        0,
        NULL,
        NULL,
        NOW() - INTERVAL '4 days',
        NOW() - INTERVAL '4 days'
    ),
    (
        '20000000-0000-7000-8000-000000000008',
        '00000000-0000-7000-8000-000000000008',
        'PAID',
        0,
        NOW() - INTERVAL '2 days 12 hours',
        NULL,
        NOW() - INTERVAL '3 days',
        NOW() - INTERVAL '2 days 12 hours'
    ),
    (
        '20000000-0000-7000-8000-000000000009',
        '00000000-0000-7000-8000-000000000009',
        'PENDING',
        0,
        NULL,
        NULL,
        NOW() - INTERVAL '2 days',
        NOW() - INTERVAL '2 days'
    ),
    (
        '20000000-0000-7000-8000-000000000010',
        '00000000-0000-7000-8000-000000000010',
        'PAID',
        0,
        NOW() - INTERVAL '12 hours',
        NULL,
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '12 hours'
    );

-- ============================================================
-- 6. ITENS DOS PEDIDOS
--
-- O preço unitário é obtido diretamente da tabela products.
-- Cada pedido possui entre 3 e 8 produtos diferentes.
-- ============================================================

INSERT INTO order_items (
    order_id,
    product_id,
    quantity,
    unit_price,
    created_at
)
SELECT
    seed_item.order_id::UUID,
    seed_item.product_id::UUID,
    seed_item.quantity,
    product.price,
    customer_order.created_at
FROM (
    VALUES
        -- Pedido 1: 4 itens
        (
            '20000000-0000-7000-8000-000000000001',
            '10000000-0000-7000-8000-000000000001',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000001',
            '10000000-0000-7000-8000-000000000002',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000001',
            '10000000-0000-7000-8000-000000000007',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000001',
            '10000000-0000-7000-8000-000000000010',
            3
        ),

        -- Pedido 2: 5 itens
        (
            '20000000-0000-7000-8000-000000000002',
            '10000000-0000-7000-8000-000000000003',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000002',
            '10000000-0000-7000-8000-000000000004',
            6
        ),
        (
            '20000000-0000-7000-8000-000000000002',
            '10000000-0000-7000-8000-000000000006',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000002',
            '10000000-0000-7000-8000-000000000015',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000002',
            '10000000-0000-7000-8000-000000000020',
            2
        ),

        -- Pedido 3: 3 itens
        (
            '20000000-0000-7000-8000-000000000003',
            '10000000-0000-7000-8000-000000000008',
            4
        ),
        (
            '20000000-0000-7000-8000-000000000003',
            '10000000-0000-7000-8000-000000000009',
            4
        ),
        (
            '20000000-0000-7000-8000-000000000003',
            '10000000-0000-7000-8000-000000000011',
            2
        ),

        -- Pedido 4: 6 itens
        (
            '20000000-0000-7000-8000-000000000004',
            '10000000-0000-7000-8000-000000000001',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000004',
            '10000000-0000-7000-8000-000000000005',
            3
        ),
        (
            '20000000-0000-7000-8000-000000000004',
            '10000000-0000-7000-8000-000000000013',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000004',
            '10000000-0000-7000-8000-000000000014',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000004',
            '10000000-0000-7000-8000-000000000017',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000004',
            '10000000-0000-7000-8000-000000000019',
            1
        ),

        -- Pedido 5: 8 itens
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000002',
            3
        ),
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000004',
            12
        ),
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000007',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000008',
            3
        ),
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000009',
            3
        ),
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000010',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000011',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000012',
            1
        ),

        -- Pedido 6: 4 itens
        (
            '20000000-0000-7000-8000-000000000006',
            '10000000-0000-7000-8000-000000000006',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000006',
            '10000000-0000-7000-8000-000000000015',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000006',
            '10000000-0000-7000-8000-000000000016',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000006',
            '10000000-0000-7000-8000-000000000018',
            1
        ),

        -- Pedido 7: 3 itens
        (
            '20000000-0000-7000-8000-000000000007',
            '10000000-0000-7000-8000-000000000003',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000007',
            '10000000-0000-7000-8000-000000000005',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000007',
            '10000000-0000-7000-8000-000000000020',
            1
        ),

        -- Pedido 8: 7 itens
        (
            '20000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000001',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000002',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000004',
            6
        ),
        (
            '20000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000007',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000008',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000010',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000012',
            2
        ),

        -- Pedido 9: 5 itens
        (
            '20000000-0000-7000-8000-000000000009',
            '10000000-0000-7000-8000-000000000009',
            5
        ),
        (
            '20000000-0000-7000-8000-000000000009',
            '10000000-0000-7000-8000-000000000011',
            4
        ),
        (
            '20000000-0000-7000-8000-000000000009',
            '10000000-0000-7000-8000-000000000013',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000009',
            '10000000-0000-7000-8000-000000000014',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000009',
            '10000000-0000-7000-8000-000000000019',
            2
        ),

        -- Pedido 10: 6 itens
        (
            '20000000-0000-7000-8000-000000000010',
            '10000000-0000-7000-8000-000000000003',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000010',
            '10000000-0000-7000-8000-000000000006',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000010',
            '10000000-0000-7000-8000-000000000015',
            3
        ),
        (
            '20000000-0000-7000-8000-000000000010',
            '10000000-0000-7000-8000-000000000016',
            2
        ),
        (
            '20000000-0000-7000-8000-000000000010',
            '10000000-0000-7000-8000-000000000017',
            1
        ),
        (
            '20000000-0000-7000-8000-000000000010',
            '10000000-0000-7000-8000-000000000018',
            2
        )
) AS seed_item (
    order_id,
    product_id,
    quantity
)
INNER JOIN products AS product
    ON product.id = seed_item.product_id::UUID
INNER JOIN orders AS customer_order
    ON customer_order.id = seed_item.order_id::UUID;

-- ============================================================
-- 7. CALCULAR O TOTAL DOS PEDIDOS
-- ============================================================

UPDATE orders AS customer_order
SET total_amount = calculated_order.total_amount
FROM (
    SELECT
        order_id,
        SUM(quantity * unit_price)::NUMERIC(12, 2) AS total_amount
    FROM order_items
    WHERE order_id IN (
        '20000000-0000-7000-8000-000000000001',
        '20000000-0000-7000-8000-000000000002',
        '20000000-0000-7000-8000-000000000003',
        '20000000-0000-7000-8000-000000000004',
        '20000000-0000-7000-8000-000000000005',
        '20000000-0000-7000-8000-000000000006',
        '20000000-0000-7000-8000-000000000007',
        '20000000-0000-7000-8000-000000000008',
        '20000000-0000-7000-8000-000000000009',
        '20000000-0000-7000-8000-000000000010'
    )
    GROUP BY order_id
) AS calculated_order
WHERE customer_order.id = calculated_order.order_id;

-- ============================================================
-- 8. ATUALIZAR ESTOQUE
--
-- PENDING e PAID já consumiram estoque.
-- CANCELED representa um pedido cujo estoque foi devolvido.
-- ============================================================

UPDATE products AS product
SET
    stock = product.stock - reserved_stock.quantity,
    updated_at = NOW()
FROM (
    SELECT
        order_item.product_id,
        SUM(order_item.quantity)::INTEGER AS quantity
    FROM order_items AS order_item
    INNER JOIN orders AS customer_order
        ON customer_order.id = order_item.order_id
    WHERE customer_order.id IN (
        '20000000-0000-7000-8000-000000000001',
        '20000000-0000-7000-8000-000000000002',
        '20000000-0000-7000-8000-000000000003',
        '20000000-0000-7000-8000-000000000004',
        '20000000-0000-7000-8000-000000000005',
        '20000000-0000-7000-8000-000000000006',
        '20000000-0000-7000-8000-000000000007',
        '20000000-0000-7000-8000-000000000008',
        '20000000-0000-7000-8000-000000000009',
        '20000000-0000-7000-8000-000000000010'
    )
      AND customer_order.status IN ('PENDING', 'PAID')
    GROUP BY order_item.product_id
) AS reserved_stock
WHERE product.id = reserved_stock.product_id;

-- ============================================================
-- 9. VALIDAÇÃO DE SEGURANÇA DO ESTOQUE
-- ============================================================

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM products
        WHERE id IN (
            '10000000-0000-7000-8000-000000000001',
            '10000000-0000-7000-8000-000000000002',
            '10000000-0000-7000-8000-000000000003',
            '10000000-0000-7000-8000-000000000004',
            '10000000-0000-7000-8000-000000000005',
            '10000000-0000-7000-8000-000000000006',
            '10000000-0000-7000-8000-000000000007',
            '10000000-0000-7000-8000-000000000008',
            '10000000-0000-7000-8000-000000000009',
            '10000000-0000-7000-8000-000000000010',
            '10000000-0000-7000-8000-000000000011',
            '10000000-0000-7000-8000-000000000012',
            '10000000-0000-7000-8000-000000000013',
            '10000000-0000-7000-8000-000000000014',
            '10000000-0000-7000-8000-000000000015',
            '10000000-0000-7000-8000-000000000016',
            '10000000-0000-7000-8000-000000000017',
            '10000000-0000-7000-8000-000000000018',
            '10000000-0000-7000-8000-000000000019',
            '10000000-0000-7000-8000-000000000020'
        )
          AND stock < 0
    ) THEN
        RAISE EXCEPTION
            'O seed tentou gerar um produto com estoque negativo';
    END IF;
END
$$;

COMMIT;

-- ============================================================
-- 10. CONSULTAS DE CONFERÊNCIA
-- ============================================================

-- Quantidade de usuários por papel.
SELECT
    role,
    COUNT(*) AS quantity
FROM users
WHERE id IN (
    '00000000-0000-7000-8000-000000000001',
    '00000000-0000-7000-8000-000000000002',
    '00000000-0000-7000-8000-000000000003',
    '00000000-0000-7000-8000-000000000004',
    '00000000-0000-7000-8000-000000000005',
    '00000000-0000-7000-8000-000000000006',
    '00000000-0000-7000-8000-000000000007',
    '00000000-0000-7000-8000-000000000008',
    '00000000-0000-7000-8000-000000000009',
    '00000000-0000-7000-8000-000000000010'
)
GROUP BY role
ORDER BY role;

-- Pedidos, proprietários, status, itens e totais.
SELECT
    customer_order.id,
    app_user.name AS customer_name,
    app_user.role AS customer_role,
    customer_order.status,
    COUNT(order_item.id) AS item_count,
    customer_order.total_amount,
    customer_order.created_at
FROM orders AS customer_order
INNER JOIN users AS app_user
    ON app_user.id = customer_order.customer_id
INNER JOIN order_items AS order_item
    ON order_item.order_id = customer_order.id
WHERE customer_order.id IN (
    '20000000-0000-7000-8000-000000000001',
    '20000000-0000-7000-8000-000000000002',
    '20000000-0000-7000-8000-000000000003',
    '20000000-0000-7000-8000-000000000004',
    '20000000-0000-7000-8000-000000000005',
    '20000000-0000-7000-8000-000000000006',
    '20000000-0000-7000-8000-000000000007',
    '20000000-0000-7000-8000-000000000008',
    '20000000-0000-7000-8000-000000000009',
    '20000000-0000-7000-8000-000000000010'
)
GROUP BY
    customer_order.id,
    app_user.name,
    app_user.role
ORDER BY customer_order.created_at;

-- Estoque final dos produtos.
SELECT
    id,
    name,
    price,
    stock
FROM products
WHERE id IN (
    '10000000-0000-7000-8000-000000000001',
    '10000000-0000-7000-8000-000000000002',
    '10000000-0000-7000-8000-000000000003',
    '10000000-0000-7000-8000-000000000004',
    '10000000-0000-7000-8000-000000000005',
    '10000000-0000-7000-8000-000000000006',
    '10000000-0000-7000-8000-000000000007',
    '10000000-0000-7000-8000-000000000008',
    '10000000-0000-7000-8000-000000000009',
    '10000000-0000-7000-8000-000000000010',
    '10000000-0000-7000-8000-000000000011',
    '10000000-0000-7000-8000-000000000012',
    '10000000-0000-7000-8000-000000000013',
    '10000000-0000-7000-8000-000000000014',
    '10000000-0000-7000-8000-000000000015',
    '10000000-0000-7000-8000-000000000016',
    '10000000-0000-7000-8000-000000000017',
    '10000000-0000-7000-8000-000000000018',
    '10000000-0000-7000-8000-000000000019',
    '10000000-0000-7000-8000-000000000020'
)
ORDER BY name;
